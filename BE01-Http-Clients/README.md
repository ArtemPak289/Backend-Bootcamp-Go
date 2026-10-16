# HTTP Public API Clients Suite & Postman Collection

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Dependencies](https://img.shields.io/badge/Dependencies-Zero%20(stdlib%20only)-success?style=for-the-badge)
![Postman](https://img.shields.io/badge/Postman-v2.1.0%20Collections-FF6C37?style=for-the-badge&logo=postman&logoColor=white)
![APIs](https://img.shields.io/badge/Integrated-NewsAPI%20%7C%20hh.ru%20%7C%20Telegram-blue?style=for-the-badge)

A production-grade Go CLI application and automated Postman collection suite demonstrating robust client-side HTTP communication, authentication patterns, JSON parsing, and multipart/form-data streaming with zero third-party dependencies (pure `net/http` standard library).

---

## 📌 Overview

This project implements robust integrations with three distinct external HTTP services, covering major web authentication and transport protocols:

1. **News API Client**:
   - Integrates with the [NewsAPI.org](https://newsapi.org/) REST API.
   - Demonstrates **Bearer Token Authentication** via HTTP headers (`Authorization: Bearer <API_KEY>`).
   - Dynamic query string construction (`q`, `sources`, `domains`, `pageSize`, `language`).
   - Graceful error decoding for rate limits (`429 Too Many Requests`) and unauthorized requests (`401 Unauthorized`).

2. **HeadHunter (hh.ru) API Client**:
   - Consumes public job search and employer endpoints from the HeadHunter API.
   - Demonstrates custom header management (`User-Agent` compliance).
   - Structured JSON deserialization for complex nested vacancy schemas, salary ranges, and employer metadata.

3. **Telegram Bot API Client**:
   - Direct integration with Telegram's Bot API (`https://api.telegram.org/bot<TOKEN>/`).
   - Implements **Multipart/Form-Data MIME encoding** via `mime/multipart` for binary image and document file uploads (`sendPhoto`, `sendDocument`).
   - Validates file MIME types to prevent erroneous submissions (e.g. attempting to upload `.pdf` or `.docx` through photo endpoints).
   - Text message formatting (`sendMessage`) and chat update polling (`getUpdates`).

---

## 🏛 Project Structure

```
BE01-Http-Clients/
├── main.go                               # CLI Application Entry Point
├── go.mod                                # Pure Go stdlib module definition
├── pkg/
│   ├── config/config.go                  # Environment variable and .env parser
│   ├── news/news.go                      # NewsAPI client (Bearer Auth & Queries)
│   ├── hh/hh.go                          # HeadHunter API client (Vacancies & Employers)
│   └── telegram/telegram.go              # Telegram client (Multipart uploads & messaging)
├── sample.jpg                            # Sample valid image for sendPhoto testing
├── sample.pdf                            # Sample document for sendDocument testing
├── sample.docx                           # Sample document for negative photo testing
├── .env.example                          # Configuration template
├── BE1_HttpAuth.postman_collection.json  # Comprehensive Postman test collection
├── BE1_HttpAuth.postman_environment.json # Postman environment variables
├── Part1_NewsAPI.postman_collection.json # Standalone NewsAPI Postman collection
├── Part2_HH.postman_collection.json      # Standalone HeadHunter Postman collection
└── Part3_TelegramBot.postman_collection.json # Standalone Telegram Postman collection
```

---

## 🚀 Getting Started

### Prerequisites
- **Go 1.20+** (no external dependencies required)
- (Optional) **Postman** or **Newman CLI** for executing API test collections

### 1. Environment Setup
Copy the configuration template:
```bash
cp .env.example .env
```

Configure your credentials inside `.env`:
```env
# NewsAPI Key (get a free key at https://newsapi.org/)
NEWS_API_KEY=your_news_api_key

# HeadHunter defaults (pre-configured with working public IDs)
HH_VACANCY_ID=138003606
HH_EMPLOYER_ID=1455

# Telegram Bot Credentials
# Create a bot via @BotFather in Telegram and obtain the token:
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
TELEGRAM_USER_ID=your_user_id
```

### 2. Running the CLI Application

The CLI tool supports running all clients or targeting individual services:

```bash
# Execute all integrations sequentially
go run . all

# Test NewsAPI integration
go run . news

# Test HeadHunter vacancy & employer search
go run . hh

# Test Telegram messaging and multipart photo uploads
go run . tg
```

---

## 📬 Postman Test Automation

The repository includes ready-to-run Postman collections with automated test assertion scripts:

- **Pre-request Scripts**: Dynamically sets timestamps and header values.
- **Tests Scripts**: Verifies HTTP `200 OK`, JSON response contract schema, and validates expected error payloads (`400 Bad Request` on invalid file extensions).

Run automated tests via **Newman**:
```bash
npm install -g newman
newman run BE1_HttpAuth.postman_collection.json -e BE1_HttpAuth.postman_environment.json
```
