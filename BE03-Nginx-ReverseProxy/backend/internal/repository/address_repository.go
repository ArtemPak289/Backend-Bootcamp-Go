package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"store_api/internal/models/dao"

	"github.com/google/uuid"
)

type AddressRepository interface {
	Create(ctx context.Context, addr *dao.AddressDAO) error
	GetByID(ctx context.Context, id uuid.UUID) (*dao.AddressDAO, error)
	FindOrCreate(ctx context.Context, country, city, street string) (*dao.AddressDAO, error)
	Update(ctx context.Context, addr *dao.AddressDAO) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresAddressRepository struct {
	db *sql.DB
}

func NewAddressRepository(db *sql.DB) AddressRepository {
	return &postgresAddressRepository{db: db}
}

func (r *postgresAddressRepository) Create(ctx context.Context, addr *dao.AddressDAO) error {
	query := `
		INSERT INTO addresses (id, country, city, street, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, addr.ID, addr.Country, addr.City, addr.Street, addr.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create address: %w", err)
	}
	return nil
}

func (r *postgresAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*dao.AddressDAO, error) {
	query := `
		SELECT id, country, city, street, created_at
		FROM addresses
		WHERE id = $1
	`
	var addr dao.AddressDAO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&addr.ID,
		&addr.Country,
		&addr.City,
		&addr.Street,
		&addr.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get address by id: %w", err)
	}
	return &addr, nil
}

func (r *postgresAddressRepository) FindOrCreate(ctx context.Context, country, city, street string) (*dao.AddressDAO, error) {
	selectQuery := `
		SELECT id, country, city, street, created_at
		FROM addresses
		WHERE LOWER(country) = LOWER($1) AND LOWER(city) = LOWER($2) AND LOWER(street) = LOWER($3)
		LIMIT 1
	`
	var addr dao.AddressDAO
	err := r.db.QueryRowContext(ctx, selectQuery, country, city, street).Scan(
		&addr.ID,
		&addr.Country,
		&addr.City,
		&addr.Street,
		&addr.CreatedAt,
	)
	if err == nil {
		return &addr, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing address: %w", err)
	}

	// Create new
	newAddr := dao.AddressDAO{
		ID:        uuid.New(),
		Country:   country,
		City:      city,
		Street:    street,
		CreatedAt: addr.CreatedAt,
	}
	if err := r.Create(ctx, &newAddr); err != nil {
		return nil, err
	}
	return &newAddr, nil
}

func (r *postgresAddressRepository) Update(ctx context.Context, addr *dao.AddressDAO) error {
	query := `
		UPDATE addresses
		SET country = $2, city = $3, street = $4
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, addr.ID, addr.Country, addr.City, addr.Street)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM addresses WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
