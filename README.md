# Go Backend Engineering & Distributed Systems Portfolio

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL%2016-336791?style=for-the-badge&logo=postgresql&logoColor=white)
![gRPC](https://img.shields.io/badge/RPC-gRPC%20%7C%20Protocol%20Buffers-244c5a?style=for-the-badge&logo=grpc&logoColor=white)
![Nginx](https://img.shields.io/badge/Proxy-Nginx%20Gateway%20%26%20Load%20Balancer-009639?style=for-the-badge&logo=nginx&logoColor=white)
![Docker Compose](https://img.shields.io/badge/Orchestration-Docker%20Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Uber Dig](https://img.shields.io/badge/DI-Uber%20Dig%20IoC-black?style=for-the-badge)
![Swagger](https://img.shields.io/badge/API%20Docs-OpenAPI%203.0%20%7C%20Swagger-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)
![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)

Welcome to my **Go Backend Engineering** portfolio repository! This monorepo demonstrates modern, scalable, and resilient backend systems developed in **Go**, covering the full spectrum of backend software engineering: from pure standard library HTTP networking to production-ready RESTful APIs, high-throughput reverse proxy gateways, and distributed **gRPC microservices**.

---

## 🏛 Repository Architecture & Project Matrix

The repository is organized into four modular, production-ready backend systems:

| Module | Category | Primary Tech Stack | Architectural Highlights |
|---|---|---|---|
| **[BE01: HTTP API Clients & Postman Suite](./BE01-Http-Clients)** | Networking & Integrations | Go 1.24 (stdlib `net/http`), Postman v2.1.0 | Zero-dependency HTTP client suite integrating with **NewsAPI** (Bearer Token Auth), **HeadHunter** (Hierarchical REST), and **Telegram Bot API** (MIME multipart/form-data streaming file uploads) with automated Postman assertion test suites. |
| **[BE02: Appliance Store RESTful API](./BE02-Store-REST-API)** | Enterprise REST Services | Go, Gorilla Mux, PostgreSQL, Swagger / OpenAPI | Clean Layered Architecture, Repository Pattern, strict DAO/DTO separation, relational schema normalized to **BCNF (Boyce-Codd Normal Form)**, connection pooling, and interactive Swagger documentation. |
| **[BE03: Nginx Gateway & Reverse Proxy](./BE03-Nginx-ReverseProxy)** | Infrastructure & DevOps | Nginx, Docker Compose, TLS 1.3, HTTP/2 | High-performance edge gateway with automated **TLS termination** (HTTP->HTTPS redirect), **method-aware load balancing** (write requests to primary; read requests load-balanced across replicas), dynamic edge HTTP caching, and Gzip compression. |
| **[BE04: gRPC Microservices Architecture](./BE04-GRPC-Microservices)** | Distributed Microservices | Go, gRPC, Protocol Buffers, Uber Dig IoC, Dual Postgres | Decoupled microservices architecture featuring an isolated **gRPC Authorization Service** and **Store REST API** with reflection-based **Dependency Injection (Uber Dig)**, cryptographic salted password hashing, and stateless JWT verification. |

---

## 🛠 Engineering Concepts & Best Practices

- **Zero-Allocation & Clean Code**: Heavy reliance on Go's standard library where appropriate (`net/http`, `mime/multipart`, `crypto/rand`, `database/sql`), writing idiomatic, concurrent, and garbage-collector-friendly code.
- **Dependency Injection (IoC Container)**: Implementation of `go.uber.org/dig` in microservices to achieve loose coupling, interface-driven software design, and high testability.
- **Distributed Communication**: High-throughput inter-service Remote Procedure Calls (RPC) powered by **Protocol Buffers** over **HTTP/2**.
- **Relational Data Modeling**: Strict adherence to normal forms (1NF, 2NF, 3NF, BCNF), `UUID` primary keys, `NUMERIC(12, 2)` monetary values to eliminate binary rounding errors, `BYTEA` binary media storage, and automated SQL migration pipelines.
- **Edge Security & Performance**: Production-ready Nginx configuration enforcing modern TLS ciphers, HTTP caching with `X-Cache-Status` observability, and read-write split load balancing.
- **Contract-Driven Development**: Full interactive API documentation using **Swagger (OpenAPI 3.0)** and automated regression test suites using **Postman & Newman**.

---

## 📂 Monorepo Structure

```text
Backend-Bootcamp-Go/
├── BE01-Http-Clients/          # Standard library HTTP clients & Postman test suite
├── BE02-Store-REST-API/        # PostgreSQL REST API with Gorilla Mux & Swagger
├── BE03-Nginx-ReverseProxy/    # Nginx Gateway, Load Balancer & TLS Reverse Proxy
├── BE04-GRPC-Microservices/    # gRPC Auth Service & REST Store API with Uber Dig
├── Makefile                    # Root automation Makefile for the entire monorepo
├── .gitignore                  # Git ignore rules for Go, Docker, and OS files
└── README.md                   # Monorepo Portfolio Documentation
```

---

## 🚀 Quickstart Guide

### Prerequisites
- **Go 1.24+** installed
- **Docker** and **Docker Compose** installed
- **Git** and **GNU Make**

---

### 1. Running BE01: HTTP Public API Clients
```bash
cd BE01-Http-Clients
cp .env.example .env
# Edit .env with your credentials

# Run CLI runner
go run . all
```

---

### 2. Running BE02: Appliance Store REST API
```bash
cd BE02-Store-REST-API
cp .env.example .env

# Run local development server
make run

# Execute integration tests
make test
```
Access the interactive Swagger UI at: `http://localhost:8080/swagger/index.html`

---

### 3. Running BE03: Nginx Gateway & Clustered Backends
```bash
cd BE03-Nginx-ReverseProxy

# Generate SSL certificates
make certs

# Launch multi-container infrastructure
make up
```
- Gateway URL: `https://localhost` (Port 80 automatically redirects to Port 443)
- pgAdmin Database GUI: `https://localhost/admin/`

---

### 4. Running BE04: Distributed gRPC Microservices
```bash
cd BE04-GRPC-Microservices

# Launch auth_service, store_service, and both PostgreSQL databases
make up

# Run end-to-end integration test suite
./scripts/test_e2e.sh

# Teardown
make down
```

---

## 📜 License

This repository is distributed under the MIT License. See individual project directories for detailed documentation, environment variable templates, and architectural notes.