package dto

// ErrorResponse represents an error response payload.
type ErrorResponse struct {
	Error string `json:"error" example:"validation failed or resource not found"`
}

// MessageResponse represents a generic success response message.
type MessageResponse struct {
	Message string `json:"message" example:"operation completed successfully"`
}
