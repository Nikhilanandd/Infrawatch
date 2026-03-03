package alert

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nikhilanandd/infrawatch/internal/config"
	"github.com/nikhilanandd/infrawatch/internal/server/store"
	"github.com/nikhilanandd/infrawatch/internal/server/websocket"
	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// Notifier is an interface for alert notification channels.
type Notifier interface {
	Send(alert models.Alert) error
	Name() string
}

// Engine evaluates alert rules against incoming metrics.
type Engine struct {
	db        *store.DB
	hub       *websocket.Hub
	notifiers []Notifier
	rules     []models.AlertRule
	interval  time.Duration
	logger    *slog.Logger
}

// NewEngine creates a new alert engine.
func NewEngine(db *store.DB, hub *websocket.Hub, cfg *config.ServerConfig, logger *slog.Logger) *Engine {
	e := &Engine{
		db:       db,
		hub:      hub,
		interval: cfg.Alerts.EvaluationInterval,
		logger:   logger,
	}

	// Seed alert rules from config
	for _, rc := range cfg.Alerts.Rules {
		rule := models.AlertRule{
			Name:      rc.Name,
			Metric:    rc.Metric,
			Operator:  rc.Operator,
			Threshold: rc.Threshold,
			Severity:  rc.Severity,
			Enabled:   rc.Enabled,
		}
		id, err := db.UpsertAlertRule(rule)
		if err != nil {
			logger.Error("failed to seed alert rule", "name", rule.Name, "error", err)
		} else {
			rule.ID = id
			e.rules = append(e.rules, rule)
		}
	}

	// Setup notifiers
	if cfg.Channels.Email.Enabled {
		e.notifiers = append(e.notifiers, NewEmailNotifier(cfg))
	}
	if cfg.Channels.Slack.Enabled {
		e.notifiers = append(e.notifiers, NewSlackNotifier(cfg))
	}

	return e
}

// Run starts the alert evaluation loop.
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	e.logger.Info("alert engine started", "interval", e.interval, "rules", len(e.rules))

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("alert engine stopped")
			return
		case <-ticker.C:
			e.evaluate()
		}
	}
}

func (e *Engine) evaluate() {
	// Refresh rules from DB
	rules, err := e.db.GetAlertRules()
	if err != nil {
		e.logger.Error("failed to get alert rules", "error", err)
		return
	}

	nodes, err := e.db.GetNodes()
	if err != nil {
		e.logger.Error("failed to get nodes", "error", err)
		return
	}

	for _, node := range nodes {
		metric, err := e.db.GetLatestMetric(node.ID)
		if err != nil {
			continue
		}

		for _, rule := range rules {
			if !rule.Enabled {
				continue
			}

			fired, value := e.checkRule(rule, metric)
			if fired {
				alert := models.Alert{
					RuleID:   rule.ID,
					NodeID:   node.ID,
					Message:  fmt.Sprintf("[%s] %s on %s: %.2f %s %.2f", rule.Severity, rule.Name, node.Hostname, value, rule.Operator, rule.Threshold),
					Severity: rule.Severity,
					Value:    value,
					FiredAt:  time.Now().UTC(),
				}

				id, err := e.db.InsertAlert(alert)
				if err != nil {
					e.logger.Error("failed to insert alert", "error", err)
					continue
				}
				alert.ID = id

				e.logger.Warn("alert fired",
					"rule", rule.Name,
					"node", node.ID,
					"value", value,
					"threshold", rule.Threshold,
				)

				// Broadcast via WebSocket
				e.hub.Broadcast(models.WSMessage{
					Type:    "alert",
					Payload: alert,
				})

				// Send notifications
				for _, n := range e.notifiers {
					if err := n.Send(alert); err != nil {
						e.logger.Error("notifier failed", "channel", n.Name(), "error", err)
					}
				}
			}
		}
	}
}

func (e *Engine) checkRule(rule models.AlertRule, m *models.MetricPayload) (bool, float64) {
	var value float64

	switch rule.Metric {
	case "cpu":
		value = m.CPU.UsagePercent
	case "disk":
		// Check the highest disk usage
		for _, p := range m.Disk.Partitions {
			if p.UsagePercent > value {
				value = p.UsagePercent
			}
		}
	case "service":
		// Check if any service is not active
		for _, s := range m.Services {
			if s.ActiveState != "active" {
				return true, 0
			}
		}
		return false, 0
	default:
		return false, 0
	}

	switch rule.Operator {
	case ">":
		return value > rule.Threshold, value
	case "<":
		return value < rule.Threshold, value
	case ">=":
		return value >= rule.Threshold, value
	case "<=":
		return value <= rule.Threshold, value
	case "==":
		return value == rule.Threshold, value
	}

	return false, value
}
