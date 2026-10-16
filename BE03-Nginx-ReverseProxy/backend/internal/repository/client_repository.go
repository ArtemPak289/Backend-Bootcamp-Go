package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"store_api/internal/models/dao"

	"github.com/google/uuid"
)

type ClientRepository interface {
	Create(ctx context.Context, client *dao.ClientDAO) error
	GetByID(ctx context.Context, id uuid.UUID) (*dao.ClientDAO, error)
	GetByNameAndSurname(ctx context.Context, firstName, lastName string) ([]dao.ClientDAO, error)
	GetAll(ctx context.Context, limit, offset *int) ([]dao.ClientDAO, error)
	UpdateAddress(ctx context.Context, id uuid.UUID, addressID uuid.UUID) (*dao.ClientDAO, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) ClientRepository {
	return &postgresClientRepository{db: db}
}

func (r *postgresClientRepository) Create(ctx context.Context, client *dao.ClientDAO) error {
	query := `
		INSERT INTO clients (id, client_name, client_surname, birthday, gender, registration_date, address_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		client.ID,
		client.ClientName,
		client.ClientSurname,
		client.Birthday,
		client.Gender,
		client.RegistrationDate,
		client.AddressID,
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}

func (r *postgresClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*dao.ClientDAO, error) {
	query := `
		SELECT id, client_name, client_surname, birthday, gender, registration_date, address_id
		FROM clients
		WHERE id = $1
	`
	var client dao.ClientDAO
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&client.ID,
		&client.ClientName,
		&client.ClientSurname,
		&client.Birthday,
		&client.Gender,
		&client.RegistrationDate,
		&client.AddressID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client by id: %w", err)
	}
	return &client, nil
}

func (r *postgresClientRepository) GetByNameAndSurname(ctx context.Context, firstName, lastName string) ([]dao.ClientDAO, error) {
	query := `
		SELECT id, client_name, client_surname, birthday, gender, registration_date, address_id
		FROM clients
		WHERE LOWER(client_name) = LOWER($1) AND LOWER(client_surname) = LOWER($2)
		ORDER BY registration_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("failed to search clients: %w", err)
	}
	defer rows.Close()

	clients := make([]dao.ClientDAO, 0)
	for rows.Next() {
		var c dao.ClientDAO
		if err := rows.Scan(
			&c.ID,
			&c.ClientName,
			&c.ClientSurname,
			&c.Birthday,
			&c.Gender,
			&c.RegistrationDate,
			&c.AddressID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan client row: %w", err)
		}
		clients = append(clients, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *postgresClientRepository) GetAll(ctx context.Context, limit, offset *int) ([]dao.ClientDAO, error) {
	query := `
		SELECT id, client_name, client_surname, birthday, gender, registration_date, address_id
		FROM clients
		ORDER BY registration_date DESC
	`
	var args []interface{}
	if limit != nil && offset != nil {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, *limit, *offset)
	} else if limit != nil {
		query += " LIMIT $1"
		args = append(args, *limit)
	} else if offset != nil {
		query += " OFFSET $1"
		args = append(args, *offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clients: %w", err)
	}
	defer rows.Close()

	clients := make([]dao.ClientDAO, 0)
	for rows.Next() {
		var c dao.ClientDAO
		if err := rows.Scan(
			&c.ID,
			&c.ClientName,
			&c.ClientSurname,
			&c.Birthday,
			&c.Gender,
			&c.RegistrationDate,
			&c.AddressID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *postgresClientRepository) UpdateAddress(ctx context.Context, id uuid.UUID, addressID uuid.UUID) (*dao.ClientDAO, error) {
	query := `
		UPDATE clients
		SET address_id = $2
		WHERE id = $1
		RETURNING id, client_name, client_surname, birthday, gender, registration_date, address_id
	`
	var client dao.ClientDAO
	err := r.db.QueryRowContext(ctx, query, id, addressID).Scan(
		&client.ID,
		&client.ClientName,
		&client.ClientSurname,
		&client.Birthday,
		&client.Gender,
		&client.RegistrationDate,
		&client.AddressID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update client address: %w", err)
	}
	return &client, nil
}

func (r *postgresClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM clients WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
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
