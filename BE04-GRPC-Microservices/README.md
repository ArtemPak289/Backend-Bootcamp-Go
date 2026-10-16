# Microservices Architecture: gRPC Auth Service & REST Store API

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![gRPC](https://img.shields.io/badge/RPC-gRPC%20%7C%20Protocol%20Buffers-244c5a?style=for-the-badge&logo=grpc&logoColor=white)
![Dependency Injection](https://img.shields.io/badge/IoC-Uber%20Dig%20Container-000000?style=for-the-badge)
![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL%20(Dual%20DB)-336791?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker Compose](https://img.shields.io/badge/Orchestration-Docker%20Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)

An enterprise-grade, distributed microservices architecture featuring a dedicated **gRPC Authorization Service** and an **Appliance Store RESTful API**. The solution demonstrates interface-based **Dependency Injection (IoC Container using Uber Dig)**, cryptographic salted password hashing, JWT authentication, and inter-service RPC communication over HTTP/2.

---

## 🏛 System Architecture

The architecture partitions responsibilities into two independent services with isolated PostgreSQL database instances:

```
                                  ┌────────────────────────┐
                                  │   HTTP Client (REST)   │
                                  └───────────┬────────────┘
                                              │ HTTP / JSON
                                              ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│  Store Service (Port 8080)                                                             │
│  - RESTful HTTP API (Gorilla Mux + Swagger OpenAPI)                                    │
│  - Custom Authorization Middleware & Claims Context Extraction                         │
│  - Dependency Injection (Uber Dig IoC Container)                                       │
│  - Public Endpoints: /register, /auth, /reset                                          │
│  - Protected Endpoints: /api/v1/clients, /api/v1/products, ...                         │
└─────────────────────────────────────────────┬──────────────────────────────────────────┘
                                              │ gRPC over HTTP/2 (Port 50051)
                                              ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│  Auth Service (Port 50051)                                                             │
│  - Pure gRPC Server implementing auth.proto IDL                                        │
│  - Dependency Injection (Uber Dig IoC Container)                                       │
│  - Salted Password Hashing (SHA-256 + 16-byte crypto salt)                             │
│  - Stateless JWT Token Issuing & Validation (HMAC-SHA256)                              │
│  - Password Recovery System                                                            │
└───────────────────────┬────────────────────────────────────────┬───────────────────────┘
                        │                                        │
                        ▼                                        ▼
             ┌─────────────────────┐                  ┌─────────────────────┐
             │ store_db (Postgres) │                  │  auth_db (Postgres) │
             │ Appliance Catalog   │                  │  Users, Salt, Hash  │
             └─────────────────────┘                  └─────────────────────┘
```

---

## 🔑 Key Architectural Highlights

### 1. Dependency Injection with Uber Dig IoC
Both microservices utilize **Uber Dig** (`go.uber.org/dig`), a reflection-based dependency injection container for Go:
- **Loose Coupling**: Every service, repository, and client depends exclusively on Go `interface` contracts.
- **Automated Lifecycle & Wiring**: The IoC container automatically resolves dependency graphs, manages singleton instances (database connection pools, gRPC connections), and injects components into HTTP handlers and RPC servers.

### 2. High-Performance gRPC Communication
- Inter-service communication between `store_service` and `auth_service` executes over **HTTP/2 with Protocol Buffers**.
- The `proto/auth.proto` definition defines clean contracts:
  - `RegisterUser(RegisterRequest) returns (RegisterResponse)`
  - `AuthenticateUser(AuthRequest) returns (AuthResponse)`
  - `ValidateToken(ValidateRequest) returns (ValidateResponse)`
  - `ResetPassword(ResetRequest) returns (ResetResponse)`

### 3. Cryptographic Security Standards
- **Salted Password Storage**: Passwords are never stored in plaintext. Each account receives a unique 16-byte random cryptographic salt generated via `crypto/rand`:
  $$\text{Stored Hash} = \text{SHA256}(\text{Password} \parallel \text{Salt})$$
- **Stateless JWT Tokens**: Signed using HMAC-SHA256 with configurable TTL. The `store_service` validates tokens on every protected HTTP request by querying the `auth_service` via gRPC.

---

## 📁 Repository Structure

```
BE04-GRPC-Microservices/
├── docker-compose.yml       # Orchestrates auth_service, store_service, auth_db, store_db
├── Makefile                 # Automation targets: build, up, down, local-up, test
├── proto/
│   └── auth.proto           # Protocol Buffer IDL contract
├── scripts/
│   ├── init_db.sql          # Database initialization scripts
│   ├── run_local.sh         # Launch all components locally in background
│   ├── stop_local.sh        # Teardown local running instances
│   └── test_e2e.sh          # Automated end-to-end integration test runner
├── auth_service/            # Pure gRPC Authentication microservice
│   ├── cmd/ / main.go       # Service entry point
│   ├── internal/
│   │   ├── di/              # Uber Dig IoC Container wiring
│   │   ├── grpcserver/      # gRPC server implementation
│   │   ├── repository/      # PostgreSQL user storage
│   │   └── security/        # Cryptographic hasher & JWT manager
│   ├── proto/               # Generated protobuf Go stubs
│   ├── schema.sql           # auth_db schema
│   └── Dockerfile
└── store_service/           # RESTful Appliance Store API
    ├── cmd/ / main.go       # Service entry point
    ├── internal/
    │   ├── di/              # Uber Dig IoC Container wiring
    │   ├── client/          # gRPC AuthClient connecting to Auth Service
    │   ├── auth/            # HTTP AuthMiddleware
    │   ├── handlers/        # REST Controllers
    │   └── repository/      # PostgreSQL store storage
    ├── docs/                # Swagger OpenAPI specifications
    ├── schema.sql           # store_db schema
    └── Dockerfile
```

---

## 📡 End-to-End API Workflow

### 1. Register a New User
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email": "engineer@example.com", "password": "SecurePassword123"}'
```

### 2. Login & Obtain JWT Token
```bash
curl -X POST http://localhost:8080/auth \
  -H "Content-Type: application/json" \
  -d '{"email": "engineer@example.com", "password": "SecurePassword123"}'
```
**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn..."
}
```

### 3. Query Protected Catalog Endpoint
```bash
curl -X GET http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>"
```

### 4. Password Reset
```bash
curl -X POST http://localhost:8080/reset \
  -H "Content-Type: application/json" \
  -d '{"email": "engineer@example.com"}'
```

---

## 🚀 How to Build & Run

### Run with Docker Compose
```bash
# Build and spin up both microservices and databases
make up

# View consolidated logs
docker compose logs -f

# Run end-to-end test suite against running containers
./scripts/test_e2e.sh

# Stop and clean up containers
make down
```

### Run Locally (Without Docker)
Ensure two PostgreSQL instances are running (ports `5432` and `5433`):
```bash
# Start both services locally in background
./scripts/run_local.sh

# Run end-to-end verification tests
./scripts/test_e2e.sh

# Stop local services
./scripts/stop_local.sh
```
