package dao

import (
	"time"

	"github.com/google/uuid"
)

// SupplierDAO represents the supplier entity in the database (DAL).
type SupplierDAO struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	AddressID   uuid.UUID `db:"address_id"`
	PhoneNumber string    `db:"phone_number"`
	CreatedAt   time.Time `db:"created_at"`
}
