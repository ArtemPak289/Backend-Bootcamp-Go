package dto

import "github.com/google/uuid"

// SupplierRequestDTO represents the payload for adding a supplier.
type SupplierRequestDTO struct {
	ID          *uuid.UUID  `json:"id,omitempty" example:"44444444-4444-4444-4444-444444444444"`
	Name        string      `json:"name" example:"Appliance Master Ltd"`
	AddressID   *uuid.UUID  `json:"address_id,omitempty" example:"22222222-2222-2222-2222-222222222222"`
	Address     *AddressDTO `json:"address,omitempty"`
	PhoneNumber string      `json:"phone_number" example:"+1-555-0199"`
}

// SupplierResponseDTO represents the returned supplier data.
type SupplierResponseDTO struct {
	ID          uuid.UUID   `json:"id" example:"44444444-4444-4444-4444-444444444444"`
	Name        string      `json:"name" example:"Appliance Master Ltd"`
	AddressID   uuid.UUID   `json:"address_id" example:"22222222-2222-2222-2222-222222222222"`
	Address     *AddressDTO `json:"address,omitempty"`
	PhoneNumber string      `json:"phone_number" example:"+1-555-0199"`
}

// SupplierAddressUpdateDTO represents the payload for changing a supplier's address.
type SupplierAddressUpdateDTO struct {
	AddressID *uuid.UUID `json:"address_id,omitempty" example:"22222222-2222-2222-2222-222222222222"`
	Country   string     `json:"country,omitempty" example:"USA"`
	City      string     `json:"city,omitempty" example:"Chicago"`
	Street    string     `json:"street,omitempty" example:"Michigan Ave, 45"`
}
