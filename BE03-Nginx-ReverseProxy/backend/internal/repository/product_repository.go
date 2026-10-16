package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"store_api/internal/models/dao"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *dao.ProductDAO) error
	GetByID(ctx context.Context, id uuid.UUID) (*dao.ProductDAO, error)
	GetAllAvailable(ctx context.Context) ([]dao.ProductDAO, error)
	GetAll(ctx context.Context) ([]dao.ProductDAO, error)
	DecreaseStock(ctx context.Context, id uuid.UUID, amount int) (*dao.ProductDAO, error)
	UpdateImageID(ctx context.Context, id uuid.UUID, imageID *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &postgresProductRepository{db: db}
}

func (r *postgresProductRepository) Create(ctx context.Context, product *dao.ProductDAO) error {
	query := `
		INSERT INTO products (id, name, category, price, available_stock, last_update_date, supplier_id, image_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.Category,
		product.Price,
		product.AvailableStock,
		product.LastUpdateDate,
		product.SupplierID,
		product.ImageID,
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *postgresProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*dao.ProductDAO, error) {
	query := `
		SELECT id, name, category, price, available_stock, last_update_date, supplier_id, image_id
		FROM products
		WHERE id = $1
	`
	var p dao.ProductDAO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Category,
		&p.Price,
		&p.AvailableStock,
		&p.LastUpdateDate,
		&p.SupplierID,
		&p.ImageID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	return &p, nil
}

func (r *postgresProductRepository) GetAllAvailable(ctx context.Context) ([]dao.ProductDAO, error) {
	query := `
		SELECT id, name, category, price, available_stock, last_update_date, supplier_id, image_id
		FROM products
		WHERE available_stock > 0
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get available products: %w", err)
	}
	defer rows.Close()

	products := make([]dao.ProductDAO, 0)
	for rows.Next() {
		var p dao.ProductDAO
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Category,
			&p.Price,
			&p.AvailableStock,
			&p.LastUpdateDate,
			&p.SupplierID,
			&p.ImageID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *postgresProductRepository) GetAll(ctx context.Context) ([]dao.ProductDAO, error) {
	query := `
		SELECT id, name, category, price, available_stock, last_update_date, supplier_id, image_id
		FROM products
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}
	defer rows.Close()

	products := make([]dao.ProductDAO, 0)
	for rows.Next() {
		var p dao.ProductDAO
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Category,
			&p.Price,
			&p.AvailableStock,
			&p.LastUpdateDate,
			&p.SupplierID,
			&p.ImageID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *postgresProductRepository) DecreaseStock(ctx context.Context, id uuid.UUID, amount int) (*dao.ProductDAO, error) {
	// First check if product exists
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.AvailableStock < amount {
		return nil, ErrInsufficientStock
	}

	query := `
		UPDATE products
		SET available_stock = available_stock - $2,
		    last_update_date = CURRENT_TIMESTAMP
		WHERE id = $1 AND available_stock >= $2
		RETURNING id, name, category, price, available_stock, last_update_date, supplier_id, image_id
	`
	var p dao.ProductDAO
	err = r.db.QueryRowContext(ctx, query, id, amount).Scan(
		&p.ID,
		&p.Name,
		&p.Category,
		&p.Price,
		&p.AvailableStock,
		&p.LastUpdateDate,
		&p.SupplierID,
		&p.ImageID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInsufficientStock
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decrease product stock: %w", err)
	}
	return &p, nil
}

func (r *postgresProductRepository) UpdateImageID(ctx context.Context, id uuid.UUID, imageID *uuid.UUID) error {
	query := `
		UPDATE products
		SET image_id = $2
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id, imageID)
	if err != nil {
		return fmt.Errorf("failed to update product image_id: %w", err)
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

func (r *postgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
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
