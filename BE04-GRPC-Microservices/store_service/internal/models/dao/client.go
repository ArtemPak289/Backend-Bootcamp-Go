package dao

import (
	"time"

	"github.com/google/uuid"
)

// ClientDAO represents the client entity in the database (DAL).
type ClientDAO struct {
	ID               uuid.UUID `db:"id"`
	ClientName       string    `db:"client_name"`
	ClientSurname    string    `db:"client_surname"`
	Birthday         time.Time `db:"birthday"`
	Gender           string    `db:"gender"`
	RegistrationDate time.Time `db:"registration_date"`
	AddressID        uuid.UUID `db:"address_id"`
}
