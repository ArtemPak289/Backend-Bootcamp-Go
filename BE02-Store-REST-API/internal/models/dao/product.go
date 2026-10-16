package dao

import (
	"time"

	"github.com/google/uuid"
)

// ProductDAO represents the product entity in the database (DAL).
type ProductDAO struct {
	ID             uuid.UUID  `db:"id"`
	Name           string     `db:"name"`
	Category       string     `db:"category"`
	Price          float64    `db:"price"`
	AvailableStock int        `db:"available_stock"`
	LastUpdateDate time.Time  `db:"last_update_date"`
	SupplierID     *uuid.UUID `db:"supplier_id"`
	ImageID        *uuid.UUID `db:"image_id"`
}
