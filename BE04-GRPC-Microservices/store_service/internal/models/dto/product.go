package dto

import "github.com/google/uuid"

// ProductRequestDTO represents the payload for adding/updating a product.
type ProductRequestDTO struct {
	ID             *uuid.UUID `json:"id,omitempty" example:"33333333-3333-3333-3333-333333333333"`
	Name           string     `json:"name" example:"Microwave Oven Pro"`
	Category       string     `json:"category" example:"Kitchen Appliances"`
	Price          float64    `json:"price" example:"299.99"`
	AvailableStock int        `json:"available_stock" example:"50"`
	LastUpdateDate *string    `json:"last_update_date,omitempty" example:"2024-01-01T00:00:00Z"`
	SupplierID     *uuid.UUID `json:"supplier_id,omitempty" example:"44444444-4444-4444-4444-444444444444"`
	ImageID        *uuid.UUID `json:"image_id,omitempty" example:"55555555-5555-5555-5555-555555555555"`
}

// ProductResponseDTO represents the returned product data.
type ProductResponseDTO struct {
	ID             uuid.UUID  `json:"id" example:"33333333-3333-3333-3333-333333333333"`
	Name           string     `json:"name" example:"Microwave Oven Pro"`
	Category       string     `json:"category" example:"Kitchen Appliances"`
	Price          float64    `json:"price" example:"299.99"`
	AvailableStock int        `json:"available_stock" example:"50"`
	LastUpdateDate string     `json:"last_update_date" example:"2024-01-01T00:00:00Z"`
	SupplierID     *uuid.UUID `json:"supplier_id,omitempty" example:"44444444-4444-4444-4444-444444444444"`
	ImageID        *uuid.UUID `json:"image_id,omitempty" example:"55555555-5555-5555-5555-555555555555"`
}

// ProductDecreaseStockDTO represents the input for decreasing product quantity.
type ProductDecreaseStockDTO struct {
	Amount     int `json:"amount,omitempty" example:"5"`
	DecreaseBy int `json:"decrease_by,omitempty" example:"5"`
	Quantity   int `json:"quantity,omitempty" example:"5"`
}
