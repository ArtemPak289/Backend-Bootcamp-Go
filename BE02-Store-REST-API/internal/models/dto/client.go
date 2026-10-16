package dto

import "github.com/google/uuid"

// ClientRequestDTO represents input for creating or updating a client.
type ClientRequestDTO struct {
	ID               *uuid.UUID  `json:"id,omitempty" example:"11111111-1111-1111-1111-111111111111"`
	ClientName       string      `json:"client_name" example:"John"`
	ClientSurname    string      `json:"client_surname" example:"Doe"`
	Birthday         string      `json:"birthday" example:"1990-05-15"`
	Gender           string      `json:"gender" example:"male"`
	RegistrationDate *string     `json:"registration_date,omitempty" example:"2024-01-01T00:00:00Z"`
	AddressID        *uuid.UUID  `json:"address_id,omitempty" example:"22222222-2222-2222-2222-222222222222"`
	Address          *AddressDTO `json:"address,omitempty"`
}

// ClientResponseDTO represents the output data for a client.
type ClientResponseDTO struct {
	ID               uuid.UUID   `json:"id" example:"11111111-1111-1111-1111-111111111111"`
	ClientName       string      `json:"client_name" example:"John"`
	ClientSurname    string      `json:"client_surname" example:"Doe"`
	Birthday         string      `json:"birthday" example:"1990-05-15"`
	Gender           string      `json:"gender" example:"male"`
	RegistrationDate string      `json:"registration_date" example:"2024-01-01T00:00:00Z"`
	AddressID        uuid.UUID   `json:"address_id" example:"22222222-2222-2222-2222-222222222222"`
	Address          *AddressDTO `json:"address,omitempty"`
}

// ClientAddressUpdateDTO represents the payload for changing a client's address.
type ClientAddressUpdateDTO struct {
	AddressID *uuid.UUID `json:"address_id,omitempty" example:"22222222-2222-2222-2222-222222222222"`
	Country   string     `json:"country,omitempty" example:"Germany"`
	City      string     `json:"city,omitempty" example:"Berlin"`
	Street    string     `json:"street,omitempty" example:"Alexanderplatz 1"`
}
