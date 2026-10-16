package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"store_api/internal/models/dao"

	"github.com/google/uuid"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *dao.SupplierDAO) error
	GetByID(ctx context.Context, id uuid.UUID) (*dao.SupplierDAO, error)
	GetAll(ctx context.Context) ([]dao.SupplierDAO, error)
	UpdateAddress(ctx context.Context, id uuid.UUID, addressID uuid.UUID) (*dao.SupplierDAO, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresSupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) SupplierRepository {
	return &postgresSupplierRepository{db: db}
}

func (r *postgresSupplierRepository) Create(ctx context.Context, supplier *dao.SupplierDAO) error {
	query := `
		INSERT INTO suppliers (id, name, address_id, phone_number, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		supplier.ID,
		supplier.Name,
		supplier.AddressID,
		supplier.PhoneNumber,
		supplier.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	return nil
}

func (r *postgresSupplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*dao.SupplierDAO, error) {
	query := `
		SELECT id, name, address_id, phone_number, created_at
		FROM suppliers
		WHERE id = $1
	`
	var s dao.SupplierDAO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.Name,
		&s.AddressID,
		&s.PhoneNumber,
		&s.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier by id: %w", err)
	}
	return &s, nil
}

func (r *postgresSupplierRepository) GetAll(ctx context.Context) ([]dao.SupplierDAO, error) {
	query := `
		SELECT id, name, address_id, phone_number, created_at
		FROM suppliers
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all suppliers: %w", err)
	}
	defer rows.Close()

	suppliers := make([]dao.SupplierDAO, 0)
	for rows.Next() {
		var s dao.SupplierDAO
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.AddressID,
			&s.PhoneNumber,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		suppliers = append(suppliers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *postgresSupplierRepository) UpdateAddress(ctx context.Context, id uuid.UUID, addressID uuid.UUID) (*dao.SupplierDAO, error) {
	query := `
		UPDATE suppliers
		SET address_id = $2
		WHERE id = $1
		RETURNING id, name, address_id, phone_number, created_at
	`
	var s dao.SupplierDAO
	err := r.db.QueryRowContext(ctx, query, id, addressID).Scan(
		&s.ID,
		&s.Name,
		&s.AddressID,
		&s.PhoneNumber,
		&s.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier address: %w", err)
	}
	return &s, nil
}

func (r *postgresSupplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM suppliers WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
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
