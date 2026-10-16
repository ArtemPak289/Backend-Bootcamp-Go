package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"store_api/internal/models/dao"

	"github.com/google/uuid"
)

type ImageRepository interface {
	Create(ctx context.Context, img *dao.ImageDAO) error
	GetByID(ctx context.Context, id uuid.UUID) (*dao.ImageDAO, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) (*dao.ImageDAO, error)
	Update(ctx context.Context, id uuid.UUID, data []byte) (*dao.ImageDAO, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) ImageRepository {
	return &postgresImageRepository{db: db}
}

func (r *postgresImageRepository) Create(ctx context.Context, img *dao.ImageDAO) error {
	query := `
		INSERT INTO images (id, image, product_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		img.ID,
		img.Image,
		img.ProductID,
		img.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	return nil
}

func (r *postgresImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*dao.ImageDAO, error) {
	query := `
		SELECT id, image, product_id, created_at
		FROM images
		WHERE id = $1
	`
	var img dao.ImageDAO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&img.ID,
		&img.Image,
		&img.ProductID,
		&img.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get image by id: %w", err)
	}
	return &img, nil
}

func (r *postgresImageRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*dao.ImageDAO, error) {
	// First try to get image via products.image_id
	queryViaProduct := `
		SELECT i.id, i.image, i.product_id, i.created_at
		FROM images i
		INNER JOIN products p ON p.image_id = i.id
		WHERE p.id = $1
		LIMIT 1
	`
	var img dao.ImageDAO
	err := r.db.QueryRowContext(ctx, queryViaProduct, productID).Scan(
		&img.ID,
		&img.Image,
		&img.ProductID,
		&img.CreatedAt,
	)
	if err == nil {
		return &img, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to query image by product id: %w", err)
	}

	// Fallback to images.product_id
	queryViaImage := `
		SELECT id, image, product_id, created_at
		FROM images
		WHERE product_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err = r.db.QueryRowContext(ctx, queryViaImage, productID).Scan(
		&img.ID,
		&img.Image,
		&img.ProductID,
		&img.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get image by product id: %w", err)
	}
	return &img, nil
}

func (r *postgresImageRepository) Update(ctx context.Context, id uuid.UUID, data []byte) (*dao.ImageDAO, error) {
	query := `
		UPDATE images
		SET image = $2
		WHERE id = $1
		RETURNING id, image, product_id, created_at
	`
	var img dao.ImageDAO
	err := r.db.QueryRowContext(ctx, query, id, data).Scan(
		&img.ID,
		&img.Image,
		&img.ProductID,
		&img.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update image: %w", err)
	}
	return &img, nil
}

func (r *postgresImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// First disassociate from products
	_, _ = r.db.ExecContext(ctx, `UPDATE products SET image_id = NULL WHERE image_id = $1`, id)

	query := `DELETE FROM images WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
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
