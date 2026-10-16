package dto

import "github.com/google/uuid"

// ImageResponseDTO represents metadata returned for an image.
type ImageResponseDTO struct {
	ID        uuid.UUID  `json:"id" example:"55555555-5555-5555-5555-555555555555"`
	ProductID *uuid.UUID `json:"product_id,omitempty" example:"33333333-3333-3333-3333-333333333333"`
	Size      int        `json:"size" example:"10240"`
}

// ImageUploadJSONDTO represents optional JSON payload for image upload.
type ImageUploadJSONDTO struct {
	ProductID uuid.UUID `json:"product_id" example:"33333333-3333-3333-3333-333333333333"`
	Image     string    `json:"image" example:"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="`
}

// ImageReplaceDTO represents the payload for replacing an image.
type ImageReplaceDTO struct {
	Image string `json:"image,omitempty" example:"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="`
	Data  string `json:"data,omitempty" example:"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="`
}
