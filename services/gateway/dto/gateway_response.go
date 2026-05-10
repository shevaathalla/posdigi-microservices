package dto

// GatewayResponse represents a standard API response
type GatewayResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}
