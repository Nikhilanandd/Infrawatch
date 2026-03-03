package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nikhilanandd/infrawatch/internal/agent/collector"
	"github.com/nikhilanandd/infrawatch/internal/agent/transport"
	"github.com/nikhilanandd/infrawatch/internal/config"
	"github.com/nikhilanandd/infrawatch/pkg/logger"
	"github.com/nikhilanandd/infrawatch/pkg/models"
)

func main() {
	cfgPath := flag.String("config", "configs/agent.yaml", "path to agent config file")
	flag.Parse()

	cfg, err := config.LoadAgentConfig(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Agent.LogLevel)
	log.Info("starting infrawatch agent",
		"node_id", cfg.Agent.NodeID,
		"interval", cfg.Agent.CollectInterval,
	)

	sender, err := transport.NewHTTPSender(
		cfg.Server.URL,
		cfg.Server.CACert,
		cfg.Server.ClientCert,
		cfg.Server.ClientKey,
		log,
	)
	if err != nil {
		log.Error("failed to create sender", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	hostname, _ := os.Hostname()

	ticker := time.NewTicker(cfg.Agent.CollectInterval)
	defer ticker.Stop()

	log.Info("agent started, collecting metrics")

	// Collect immediately on start, then on ticker
	collect(ctx, cfg, log, sender, hostname)
	for {
		select {
		case <-ctx.Done():
			log.Info("agent stopped")
			return
		case <-ticker.C:
			collect(ctx, cfg, log, sender, hostname)
		}
	}
}

func collect(ctx context.Context, cfg *config.AgentConfig, log *slog.Logger, sender *transport.HTTPSender, hostname string) {
	payload := &models.MetricPayload{
		NodeID:    cfg.Agent.NodeID,
		Hostname:  hostname,
		Timestamp: time.Now().UTC(),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	// CPU
	if cfg.Collectors.CPU {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cpu, err := collector.CollectCPU()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				log.Warn("cpu collection failed", "error", err)
				return
			}
			payload.CPU = cpu
		}()
	}

	// Memory
	if cfg.Collectors.Memory {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mem, err := collector.CollectMemory()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				log.Warn("memory collection failed", "error", err)
				return
			}
			payload.Memory = mem
		}()
	}

	// Disk
	if cfg.Collectors.Disk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			disk, err := collector.CollectDisk()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				log.Warn("disk collection failed", "error", err)
				return
			}
			payload.Disk = disk
		}()
	}

	// Network
	if cfg.Collectors.Network {
		wg.Add(1)
		go func() {
			defer wg.Done()
			net, err := collector.CollectNetwork()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				log.Warn("network collection failed", "error", err)
				return
			}
			payload.Network = net
		}()
	}

	// Docker
	if cfg.Collectors.Docker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			containers, err := collector.CollectDocker()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				log.Debug("docker collection skipped", "error", err)
				return
			}
			payload.Containers = containers
		}()
	}

	// Systemd
	if cfg.Collectors.Systemd && len(cfg.Services) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svcs, err := collector.CollectSystemd(cfg.Services)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				log.Warn("systemd collection failed", "error", err)
				return
			}
			payload.Services = svcs
		}()
	}

	// Uptime
	if cfg.Collectors.Uptime {
		wg.Add(1)
		go func() {
			defer wg.Done()
			uptime, err := collector.CollectUptime()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				log.Warn("uptime collection failed", "error", err)
				return
			}
			payload.Uptime = uptime
		}()
	}

	wg.Wait()

	log.Info("metrics collected",
		"cpu", payload.CPU.UsagePercent,
		"mem", payload.Memory.UsagePercent,
		"containers", len(payload.Containers),
		"services", len(payload.Services),
	)

	if err := sender.Send(ctx, payload); err != nil {
		log.Error("failed to send metrics", "error", err)
	}
}
