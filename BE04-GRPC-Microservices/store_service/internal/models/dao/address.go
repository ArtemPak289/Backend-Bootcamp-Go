package dao

import (
	"time"

	"github.com/google/uuid"
)

// AddressDAO represents the address entity in the database (DAL).
type AddressDAO struct {
	ID        uuid.UUID `db:"id"`
	Country   string    `db:"country"`
	City      string    `db:"city"`
	Street    string    `db:"street"`
	CreatedAt time.Time `db:"created_at"`
}
