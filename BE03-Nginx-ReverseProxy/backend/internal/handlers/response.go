package handlers

import (
	"encoding/json"
	"net/http"
	"store_api/internal/models/dto"

	"github.com/google/uuid"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, dto.ErrorResponse{Error: message})
}

func parseUUID(param string) (uuid.UUID, error) {
	return uuid.Parse(param)
}
