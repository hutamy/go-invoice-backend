# Go Invoice Backend

## Features

- **User Authentication** (JWT)
- **Client Management** (CRUD)
- **Invoice Management** (CRUD)
- **PDF Invoice Generation** using HTML templates
- **Swagger/OpenAPI Docs**
- **Public Invoice Generator** (no login, instant PDF generation without data storage)

## Setup

### 1. Clone the repo

```bash
git clone https://github.com/hutamy/go-invoice-backend
cd go-invoice-backend
```

### 2. Set up .env

```
cp .env.example .env
# fill in DB, JWT_SECRET
```

### 3. Run with docker compose

```
docker-compose up --build
```

## API Documentation

Visit: `http://localhost:8080/swagger/index.html`

## 💡 Project Structure

```
├── cmd/                  # Main application entrypoint
├── config/               # Configuration files and helpers
├── docs/                 # Swagger/OpenAPI docs
├── internal/             # Internal packages
├── middleware/           # Middleware packages
├── pkg/                  # External packages
├── .env.example
├── .gitignore
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## How It Works

- **Public Mode**: Anyone can POST invoice data and receive a PDF (no auth required)
- **Authenticated Mode**: Logged-in users can save clients, manage invoices, and view history
