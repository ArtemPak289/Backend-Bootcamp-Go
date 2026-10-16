package dto

import "github.com/google/uuid"

// AddressDTO represents an address data transfer object.
type AddressDTO struct {
	ID      *uuid.UUID `json:"id,omitempty" example:"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"`
	Country string     `json:"country" example:"USA"`
	City    string     `json:"city" example:"New York"`
	Street  string     `json:"street" example:"5th Avenue, 10"`
}
