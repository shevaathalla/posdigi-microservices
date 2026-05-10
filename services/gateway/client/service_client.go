package client

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ServiceClient handles communication with backend services
type ServiceClient struct {
	baseURL    string
	httpClient *http.Client
	serviceKey string
	logger     *logrus.Logger
}

// NewServiceClient creates a new service client
func NewServiceClient(baseURL string, serviceKey string, logger *logrus.Logger) *ServiceClient {
	return &ServiceClient{
		baseURL:    baseURL,
		serviceKey: serviceKey,
		logger:     logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				// Disable auto-compression: the gateway is a transparent proxy.
				// Backends should return plain JSON; the gateway's own Gzip middleware
				// handles compression for the final client response.
				DisableCompression: true,
			},
		},
	}
}

// ForwardRequest forwards an HTTP request to the backend service
func (sc *ServiceClient) ForwardRequest(method, path string, headers map[string]string, body []byte) (*http.Response, error) {
	// Build target URL
	url := sc.baseURL + path

	sc.logger.WithFields(logrus.Fields{
		"method": method,
		"url":    url,
	}).Debug("Forwarding request to backend service")

	// Create request
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		sc.logger.WithError(err).Error("Failed to create request")
		return nil, err
	}

	// Copy headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add internal service authentication
	req.Header.Set("X-Service-Auth", sc.serviceKey)

	// Send request
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		sc.logger.WithError(err).Error("Failed to send request to backend service")
		return nil, err
	}

	return resp, nil
}

// ForwardRequestWithContext forwards a request and returns response with body
func (sc *ServiceClient) ForwardRequestWithContext(method, path string, headers map[string]string, body []byte) (int, map[string]string, []byte, error) {
	resp, err := sc.ForwardRequest(method, path, headers, body)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		sc.logger.WithError(err).Error("Failed to read response body")
		return 0, nil, nil, err
	}

	// Copy response headers
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}

	return resp.StatusCode, respHeaders, respBody, nil
}
