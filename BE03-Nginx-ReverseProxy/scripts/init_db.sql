-- Database Schema & Initialization Script for Appliance Store REST API
-- Normal Forms: 1NF, 2NF, 3NF, BCNF compliant

CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY,
    country VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    street VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS suppliers (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address_id UUID NOT NULL REFERENCES addresses(id) ON DELETE RESTRICT,
    phone_number VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY,
    client_name VARCHAR(100) NOT NULL,
    client_surname VARCHAR(100) NOT NULL,
    birthday DATE NOT NULL,
    gender VARCHAR(20) NOT NULL,
    registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    address_id UUID NOT NULL REFERENCES addresses(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY,
    image BYTEA NOT NULL,
    product_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    price NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    available_stock INTEGER NOT NULL DEFAULT 0 CHECK (available_stock >= 0),
    last_update_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    image_id UUID REFERENCES images(id) ON DELETE SET NULL
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_images_product'
    ) THEN
        ALTER TABLE images ADD CONSTRAINT fk_images_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_clients_name_surname ON clients (client_name, client_surname);
CREATE INDEX IF NOT EXISTS idx_products_available_stock ON products (available_stock);
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);
CREATE INDEX IF NOT EXISTS idx_products_supplier ON products (supplier_id);

-- Create Read-Only User for load-balanced replicas
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'store_readonly') THEN
        CREATE ROLE store_readonly WITH LOGIN PASSWORD 'readonly_pass';
    END IF;
END $$;

-- Grant permissions to read-only user
GRANT CONNECT ON DATABASE store_db TO store_readonly;
GRANT USAGE ON SCHEMA public TO store_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO store_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO store_readonly;
REVOKE INSERT, UPDATE, DELETE, TRUNCATE ON ALL TABLES IN SCHEMA public FROM store_readonly;

-- Seed Sample Data
INSERT INTO addresses (id, country, city, street)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Russian Federation', 'Moscow', 'Tverskaya St, 12'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Russian Federation', 'Saint Petersburg', 'Nevsky Prospekt, 45')
ON CONFLICT (id) DO NOTHING;

INSERT INTO suppliers (id, name, address_id, phone_number)
VALUES 
    ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'ElectroTech Global', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '+7 (495) 123-45-67')
ON CONFLICT (id) DO NOTHING;

INSERT INTO clients (id, client_name, client_surname, birthday, gender, address_id)
VALUES 
    ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Ivan', 'Petrov', '1990-05-15', 'male', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22')
ON CONFLICT (id) DO NOTHING;

INSERT INTO products (id, name, category, price, available_stock, supplier_id)
VALUES 
    ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'Smart Refrigerator Pro 5000', 'Refrigerators', 899.99, 15, 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33'),
    ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'Microwave WaveMaster 900W', 'Microwaves', 149.50, 42, 'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33')
ON CONFLICT (id) DO NOTHING;
