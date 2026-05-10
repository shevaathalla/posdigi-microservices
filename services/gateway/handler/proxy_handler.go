package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"posdigi-gateway/client"
)

// ProxyHandler handles request routing and proxying to backend services
type ProxyHandler struct {
	authClient       *client.ServiceClient
	userClient       *client.ServiceClient
	attendanceClient *client.ServiceClient
	logger           *logrus.Logger
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(authClient, userClient, attendanceClient *client.ServiceClient, logger *logrus.Logger) *ProxyHandler {
	return &ProxyHandler{
		authClient:       authClient,
		userClient:       userClient,
		attendanceClient: attendanceClient,
		logger:           logger,
	}
}

// ProxyToAuth proxies requests to the Auth service
func (ph *ProxyHandler) ProxyToAuth(c echo.Context) error {
	return ph.proxyRequest(c, ph.authClient)
}

// ProxyToUser proxies requests to the User service
func (ph *ProxyHandler) ProxyToUser(c echo.Context) error {
	return ph.proxyRequest(c, ph.userClient)
}

// ProxyToAttendance proxies requests to the Attendance service
func (ph *ProxyHandler) ProxyToAttendance(c echo.Context) error {
	return ph.proxyRequest(c, ph.attendanceClient)
}

// proxyRequest forwards the request to the appropriate backend service
func (ph *ProxyHandler) proxyRequest(c echo.Context, serviceClient *client.ServiceClient) error {
	// Extract request details
	method := c.Request().Method
	path := c.Request().URL.Path

	// Read request body
	var body []byte
	if c.Request().Body != nil {
		var err error
		body, err = io.ReadAll(c.Request().Body)
		if err != nil {
			ph.logger.WithError(err).Error("Failed to read request body")
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to read request body",
			})
		}
	}

	// Copy headers (exclude hop-by-hop headers)
	headers := make(map[string]string)
	for key, values := range c.Request().Header {
		if !isHopByHopHeader(key) {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
	}

	// Forward request to backend service
	statusCode, respHeaders, respBody, err := serviceClient.ForwardRequestWithContext(method, path, headers, body)
	if err != nil {
		ph.logger.WithError(err).Error("Failed to forward request to backend service")
		return c.JSON(http.StatusBadGateway, map[string]interface{}{
			"success": false,
			"message": "Failed to reach backend service",
		})
	}

	// Set response headers (hop-by-hop headers already excluded above)
	for key, value := range respHeaders {
		if !isHopByHopHeader(key) {
			c.Response().Header().Set(key, value)
		}
	}

	// Determine content type from backend response
	contentType := respHeaders["Content-Type"]
	if contentType == "" {
		contentType = "application/json"
	}

	// Write response using Echo's Blob — respects middleware chain properly
	return c.Blob(statusCode, contentType, respBody)
}

// isHopByHopHeader checks if a header is hop-by-hop (should not be forwarded)
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
		// Strip encoding/length headers — the gateway re-encodes via its own Gzip
		// middleware, so forwarding these from backends causes double-encoding bugs
		"Content-Encoding",
		"Content-Length",
	}

	header = strings.ToLower(header)
	for _, h := range hopByHopHeaders {
		if strings.ToLower(h) == header {
			return true
		}
	}
	return false
}
