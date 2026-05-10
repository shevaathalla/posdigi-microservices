package service

import (
	"net/http"
	"sync"
	"time"

	"posdigi-gateway/config"

	"github.com/sirupsen/logrus"
)

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Healthy      bool      `json:"healthy"`
	LastCheck    time.Time `json:"last_check"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// HealthChecker monitors the health of backend services
type HealthChecker struct {
	config        *config.Config
	logger        *logrus.Logger
	healthStatus  map[string]*ServiceHealth
	mu            sync.RWMutex
	httpClient    *http.Client
	stopChan      chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.Config, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		config: cfg,
		logger: logger,
		healthStatus: make(map[string]*ServiceHealth),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins the health checking process
func (hc *HealthChecker) Start() {
	hc.logger.Info("Starting health checker...")

	// Initialize health status
	hc.mu.Lock()
	hc.healthStatus["auth"] = &ServiceHealth{Healthy: false, LastCheck: time.Now()}
	hc.healthStatus["user"] = &ServiceHealth{Healthy: false, LastCheck: time.Now()}
	hc.healthStatus["attendance"] = &ServiceHealth{Healthy: false, LastCheck: time.Now()}
	hc.mu.Unlock()

	// Start background health checker
	go hc.run()
}

// Stop stops the health checking process
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	hc.logger.Info("Health checker stopped")
}

// run performs periodic health checks
func (hc *HealthChecker) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Perform initial health check
	hc.checkAllServices()

	for {
		select {
		case <-ticker.C:
			hc.checkAllServices()
		case <-hc.stopChan:
			return
		}
	}
}

// checkAllServices checks the health of all backend services
func (hc *HealthChecker) checkAllServices() {
	hc.checkServiceHealth("auth", hc.config.AuthServiceURL+"/health")
	hc.checkServiceHealth("user", hc.config.UserServiceURL+"/health")
	hc.checkServiceHealth("attendance", hc.config.AttendanceServiceURL+"/health")
}

// checkServiceHealth checks the health of a single service
func (hc *HealthChecker) checkServiceHealth(serviceName, healthURL string) {
	resp, err := hc.httpClient.Get(healthURL)
	healthy := err == nil && resp.StatusCode == http.StatusOK
	if resp != nil {
		resp.Body.Close()
	}

	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.healthStatus[serviceName] = &ServiceHealth{
		Healthy:      healthy,
		LastCheck:    time.Now(),
		ErrorMessage: hc.getErrorMessage(err, resp),
	}

	if !healthy {
		hc.logger.WithFields(logrus.Fields{
			"service": serviceName,
			"error":   hc.getErrorMessage(err, resp),
		}).Warn("Service health check failed")
	}
}

// getErrorMessage generates an error message for health check failures
func (hc *HealthChecker) getErrorMessage(err error, resp *http.Response) string {
	if err != nil {
		return err.Error()
	}
	if resp == nil {
		return "No response"
	}
	if resp.StatusCode != http.StatusOK {
		return "Unexpected status code: " + resp.Status
	}
	return ""
}

// GetHealthStatus returns the current health status of all services
func (hc *HealthChecker) GetHealthStatus() map[string]*ServiceHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	status := make(map[string]*ServiceHealth)
	for k, v := range hc.healthStatus {
		status[k] = &ServiceHealth{
			Healthy:      v.Healthy,
			LastCheck:    v.LastCheck,
			ErrorMessage: v.ErrorMessage,
		}
	}
	return status
}

// IsAllHealthy checks if all services are healthy
func (hc *HealthChecker) IsAllHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	for _, status := range hc.healthStatus {
		if !status.Healthy {
			return false
		}
	}
	return true
}
