package dao

import (
	"time"

	"github.com/google/uuid"
)

// ImageDAO represents the image entity in the database (DAL).
type ImageDAO struct {
	ID        uuid.UUID  `db:"id"`
	Image     []byte     `db:"image"`
	ProductID *uuid.UUID `db:"product_id"`
	CreatedAt time.Time  `db:"created_at"`
}
