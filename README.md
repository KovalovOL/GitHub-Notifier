# GitHub Notifier

A Go-based service that monitors GitHub repositories for new tag/version releases and automatically sends email notifications to subscribers. It runs a lightweight background scanner and exposes a clean REST API.

## Swagger documentation
You can find the Swagger documentation at: [https://gh-notifier.online/docs](https://gh-notifier.online/docs)

## Features

- **Background Scanning**: Periodically polls the GitHub API for new tags.
- **Transactional Emails**: Sends verification codes, subscription confirmations, and new release alerts.
- **Robust Database Schema**: Tracks user subscriptions with unique constraints and state verification.
- **API Documentation**: Built-in Swagger UI for testing endpoints.
- **Production Ready**: Bundled with Docker, Docker Compose, and custom Traefik reverse proxy configuration.

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/subscribe` | `POST` | Request subscription to a repository (sends a verification email) |
| `/api/confirm` | `GET` | Confirm the subscription using the UUID token sent via email |
| `/api/subscriptions` | `GET` | List all active/pending subscriptions for an email address |
| `/api/unsubscribe` | `GET` | Instantly unsubscribe and delete subscription using the token |
| `/docs` / `/swagger/index.html` | `GET` | View interactive Swagger API documentation |

---

## Getting Started

### Prerequisites

- Go 1.25+ (if running locally without Docker)
- Docker and Docker Compose
- PostgreSQL database
- A GitHub Personal Access Token (for API rate limits)
- SMTP credentials (e.g., Gmail App Password)

### Environment Setup

Create a `.env` file in the root directory:

```ini
DB_USER=user
DB_PASSWORD=password
DB_NAME=my_database
DB_HOST=localhost
DB_PORT=5432

GH_TOKEN=your_github_personal_access_token

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_smtp_app_password
SMTP_FROM=your_email@gmail.com
```

### Running the App

#### Via Docker Compose (Recommended)
This compiles the Go app in a multi-stage Alpine build, provisions a PostgreSQL container, sets up a network bridge, runs database migrations automatically, and starts the API:

```bash
docker compose up -d --build
```

#### Via Local Go Binary
Ensure you have a local PostgreSQL instance running matching your `.env` configuration:

```bash
# Run database migrations
make migrate-up

# Start the Go server
go run cmd/api/main.go
```

---

## Example API Usage Flow

### 1. Request a subscription
Subscribe a user email to track `KovalovOL/GitHub-Notifier`:

```bash
curl -X POST https://gh-notifier.online/api/subscribe \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "repo": "KovalovOL/GitHub-Notifier"}'
```

### 2. Confirm the subscription
After registering, the user receives an email containing a unique confirmation token. Confirm it by sending a `GET` request:

```bash
curl -X GET "https://gh-notifier.online/api/confirm?token=YOUR_UUID_TOKEN"
```

### 3. List active subscriptions
Check all tracked repositories for a specific email address:

```bash
curl -X GET "https://gh-notifier.online/api/subscriptions?email=user@example.com"
```

### 4. Unsubscribe
To stop receiving release notifications, send a `GET` request using the unsubscribe link/token:

```bash
curl -X GET "https://gh-notifier.online/api/unsubscribe?token=YOUR_UUID_TOKEN"
```