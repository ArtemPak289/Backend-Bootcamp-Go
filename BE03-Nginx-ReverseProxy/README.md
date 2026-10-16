# Nginx Gateway, Reverse Proxy & Load Balancer

![Nginx](https://img.shields.io/badge/Nginx-Reverse%20Proxy%20%26%20Gateway-009639?style=for-the-badge&logo=nginx&logoColor=white)
![Docker Compose](https://img.shields.io/badge/Orchestration-Docker%20Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![TLS 1.3](https://img.shields.io/badge/Security-TLS%201.3%20%7C%20HTTPS-red?style=for-the-badge)
![Go](https://img.shields.io/badge/Backend-Go%20REST%20Cluster-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white)

A production-ready infrastructure gateway built with **Nginx** and orchestrated via **Docker Compose**. It acts as an edge reverse proxy, weighted load balancer, TLS termination point, HTTP cache, and static asset distribution engine for a clustered Appliance Store REST API.

---

## 🏛 System Architecture

```
                          Client / Web Browser / curl
                                       │
                      ┌────────────────┴────────────────┐
                      │                                 │
             HTTP Port 80                      HTTPS Port 443
         (301 Redirect to HTTPS)            (TLSv1.2 / TLSv1.3, HTTP/2)
                      │                                 │
                      └────────────────┬────────────────┘
                                       ▼
                      ┌─────────────────────────────────┐
                      │          NGINX GATEWAY          │
                      │  - SSL Termination              │
                      │  - HTTP Caching (proxy_cache)   │
                      │  - Gzip Compression             │
                      │  - Method-Aware Routing         │
                      └────────────────┬────────────────┘
                                       │
           ┌───────────────────────────┼───────────────────────────┐
           │                           │                           │
           ▼                           ▼                           ▼
      Static Assets              API Upstreams               pgAdmin Web GUI
     (/index.html &             (/api & /api/v1)              (/admin path)
       image.png)                      │                        Port 5050
                                       │
                   ┌───────────────────┼───────────────────┐
                   │ (Weight 2)        │ (Weight 1)        │ (Weight 1)
                   ▼                   ▼                   ▼
             Backend Primary     Backend Replica 1   Backend Replica 2
              (Read-Write)          (Read-Only)         (Read-Only)
               Port 8081            Port 8082           Port 8083
                   │                   │                   │
                   └───────────────────┼───────────────────┘
                                       ▼
                        ┌──────────────────────────────┐
                        │     PostgreSQL Database      │
                        │          (store_db)          │
                        └──────────────────────────────┘
```

---

## ⚡ Core Capabilities

### 1. TLS Termination & Modern HTTPS Security
- Listens on Port 80 and Port 443 with automated `301 Moved Permanently` redirects for all insecure traffic.
- Enforces modern cryptographic cipher suites under **TLSv1.2** and **TLSv1.3**.
- Native **HTTP/2** protocol multiplexing for low-latency transmission.
- Includes self-signed SSL certificate generation scripts (`make certs`) with complete SAN (Subject Alternative Name) support for `localhost`.

### 2. Method-Aware Load Balancing
- **Write Operations** (`POST`, `PUT`, `PATCH`, `DELETE`): Automatically directed exclusively to the **Primary Read-Write Backend** to prevent database replica conflicts.
- **Read Operations** (`GET`, `HEAD`): Distributed across the backend cluster using weighted round-robin (`weight=2` for primary, `weight=1` for read replicas).

### 3. Edge HTTP Caching & Compression
- In-memory and disk-backed `proxy_cache` storing responses from idempotent `GET` queries.
- Injects `X-Cache-Status` response header (`HIT`, `MISS`, `BYPASS`) for transparent observability.
- Gzip compression dynamically optimizes payload sizes for JSON, HTML, CSS, and SVG responses.

### 4. Reverse Proxying & Admin Routing
- Proxies `/admin/` traffic transparently to **pgAdmin 4**, enabling direct browser-based database management.
- High-efficiency static file delivery utilizing `sendfile on;` and `tcp_nopush on;`.

---

## 📁 Repository Structure

```
BE03-Nginx-ReverseProxy/
├── docker-compose.yml       # Multi-container orchestration (Nginx, 3x Go Backends, Postgres, pgAdmin)
├── Makefile                 # Automation targets: build, up, down, test, certs
├── nginx/
│   ├── nginx.conf           # Global Nginx server configuration & cache definition
│   └── conf.d/              # Virtual host configurations & upstream clusters
├── ssl/                     # TLS certificate (server.crt) and private key (server.key)
├── static/                  # Static website content (index.html, image.png)
├── scripts/                 # Health check and validation test scripts
└── backend/                 # Appliance Store Go REST API source code
```

---

## 🚀 Getting Started

### Prerequisites
- **Docker** & **Docker Compose**
- **GNU Make** (optional, for automation)

### 1. Generate SSL Certificates
```bash
make certs
```

### 2. Launch the Entire Cluster
```bash
# Start all containers in detached mode
make up
```

This starts:
- `nginx_gateway` on `https://localhost:443` (and HTTP `http://localhost:80`)
- `backend_primary` on internal port `8081`
- `backend_replica_1` on internal port `8082`
- `backend_replica_2` on internal port `8083`
- `postgres_db` on port `5432`
- `pgadmin_gui` accessible via `https://localhost/admin/`

### 3. Verify Endpoints

#### HTTPS & Redirect Validation:
```bash
# Check automatic HTTP -> HTTPS redirect
curl -I http://localhost/

# Query API through HTTPS
curl -k https://localhost/api/v1/products
```

#### Inspect Caching Headers:
```bash
curl -k -I https://localhost/api/v1/products
# Response header: X-Cache-Status: HIT
```

### 4. Stop Cluster
```bash
make down
```
