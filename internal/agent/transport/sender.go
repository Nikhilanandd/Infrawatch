package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// HTTPSender sends metric payloads to the central server over HTTPS with mTLS.
type HTTPSender struct {
	client   *http.Client
	endpoint string
	logger   *slog.Logger
}

// NewHTTPSender creates a new mTLS-enabled HTTP sender.
func NewHTTPSender(endpoint, caCertPath, clientCertPath, clientKeyPath string, logger *slog.Logger) (*HTTPSender, error) {
	tlsCfg := &tls.Config{}

	// Load CA cert for server verification
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert")
		}
		tlsCfg.RootCAs = pool
	}

	// Load client certificate for mTLS
	if clientCertPath != "" && clientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load client cert: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	// If no certs provided (dev mode), skip verification
	if caCertPath == "" {
		tlsCfg.InsecureSkipVerify = true
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: 10 * time.Second,
	}

	return &HTTPSender{
		client:   client,
		endpoint: endpoint,
		logger:   logger,
	}, nil
}

// Send transmits a MetricPayload to the server.
func (s *HTTPSender) Send(ctx context.Context, payload *models.MetricPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	s.logger.Debug("metrics sent successfully", "status", resp.StatusCode, "bytes", len(data))
	return nil
}
