# Message Sending Service

A robust automatic message sending system built with Go, PostgreSQL, and Redis. This service automatically sends messages retrieved from the database that have not yet been sent, processing 2 messages every 2 minutes through a configurable scheduler.

## 🚀 Features

- **Automatic Message Scheduling**: Sends 2 messages every 2 minutes automatically
- **Message Management**: Create, retrieve, and track message status
- **External API Integration**: Sends messages via configurable external API endpoints
- **Redis Caching**: Caches sent message information (messageId + sending time)
- **RESTful API**: Complete REST API with Swagger documentation
- **Scheduler Control**: Start/stop automatic message sending via API
- **Database Support**: PostgreSQL with proper indexing and constraints
- **Clean Architecture**: Follows domain-driven design principles
- **Docker Support**: Containerized deployment with docker-compose
- **Comprehensive Logging**: Structured logging with Zap
- **Health Checks**: Built-in health check endpoints

## 📋 Requirements

- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose (optional)

## 🚀 Quick Start

Get the service running in under 2 minutes:

```bash
# Clone and start
git clone https://github.com/artista-qrue/message-sending-service.git
cd message-sending-service
docker compose up -d

# Test the API
curl http://localhost:8080/health

# View Swagger docs
open http://localhost:8080/swagger/index.html
```

## 🛠️ Installation

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/artista-qrue/message-sending-service.git
   cd message-sending-service
   ```

2. **Configure environment variables**
   ```bash
   cp config.env.example config.env
   # Edit config.env with your settings
   ```

3. **Start the services**
   ```bash
   docker compose up -d
   ```

4. **Verify services are running**
   ```bash
   docker compose ps
   ```

   The database is automatically initialized with the schema from `scripts/init.sql`.

### Manual Installation

1. **Install dependencies**
   ```bash
   # Install PostgreSQL
   brew install postgresql
   
   # Install Redis
   brew install redis
   
   # Start services
   brew services start postgresql
   brew services start redis
   ```

2. **Setup database**
   ```bash
   # Create user and database
   createuser -s message_user
   createdb -O message_user message_db
   
   # Run initialization script
   psql -U message_user -d message_db -f scripts/init.sql
   ```

3. **Configure environment**
   ```bash
   cp config.env.example config.env
   # Edit config.env with your database credentials
   ```

4. **Install Go dependencies**
   ```bash
   go mod download
   ```

5. **Build and run**
   ```bash
   make build
   make run
   # OR
   go run cmd/server/main.go
   ```

## ⚙️ Configuration

The application is configured via environment variables in `config.env`:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=message_user
DB_PASSWORD=message123
DB_NAME=message_db
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# External Message API Configuration
MESSAGE_API_URL=https://webhook.site/YOUR_WEBHOOK_ID
MESSAGE_API_TIMEOUT=30s

# Scheduler Configuration
SCHEDULER_INTERVAL=2m
MESSAGES_PER_BATCH=2

# Logging
LOG_LEVEL=info
```

### External API Setup

For testing, you can use webhook.site:

1. Go to https://webhook.site
2. Copy your unique URL
3. Update `MESSAGE_API_URL` in `config.env`

The external API expects POST requests with this format:
```json
{
  "phone_number": "+1234567890",
  "message": "Your message content"
}
```

## 📖 API Documentation

Once the service is running, access the Swagger documentation at:
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **JSON**: http://localhost:8080/swagger/doc.json
- **YAML**: http://localhost:8080/swagger/swagger.yaml

### Key Endpoints

#### Messages
- `POST /api/v1/messages` - Create a new message
- `GET /api/v1/messages/{id}` - Get message by ID
- `GET /api/v1/messages/sent` - Get list of sent messages
- `GET /api/v1/messages/stats` - Get message statistics
- `POST /api/v1/messages/{id}/send` - Send specific message

#### Scheduler
- `POST /api/v1/scheduler/start` - Start automatic sending
- `POST /api/v1/scheduler/stop` - Stop automatic sending
- `GET /api/v1/scheduler/status` - Get scheduler status

#### Health Check
- `GET /health` - Service health check

## 🔄 Usage Examples

### Complete Demo Workflow

```bash
# 1. Create a test message
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, this is a test message!",
    "phone_number": "+1234567890"
  }'

# 2. Check scheduler status (should be running by default)
curl http://localhost:8080/api/v1/scheduler/status

# 3. Get message statistics
curl http://localhost:8080/api/v1/messages/stats

# 4. Send a specific message manually
curl -X POST http://localhost:8080/api/v1/messages/{message-id}/send

# 5. View sent messages with pagination
curl "http://localhost:8080/api/v1/messages/sent?page=1&limit=10"

# 6. Stop automatic sending if needed
curl -X POST http://localhost:8080/api/v1/scheduler/stop

# 7. Restart automatic sending
curl -X POST http://localhost:8080/api/v1/scheduler/start
```

### Individual API Calls

#### Create a Message
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, this is a test message!",
    "phone_number": "+1234567890"
  }'
```

#### Start Automatic Sending
```bash
curl -X POST http://localhost:8080/api/v1/scheduler/start
```

#### Get Scheduler Status
```bash
curl http://localhost:8080/api/v1/scheduler/status
```

#### Get Sent Messages
```bash
curl "http://localhost:8080/api/v1/messages/sent?page=1&limit=10"
```

## 🏗️ Architecture

The project follows Clean Architecture principles:

```
cmd/                    # Application entry points
internal/
├── domain/            # Business logic and entities
│   ├── entities/      # Domain entities
│   ├── repositories/  # Repository interfaces
│   └── usecases/      # Use case interfaces
├── application/       # Application services
│   ├── dto/          # Data transfer objects
│   ├── handlers/     # HTTP handlers
│   ├── middlewares/  # HTTP middlewares
│   └── usecases/     # Use case implementations
├── infrastructure/   # External dependencies
│   ├── config/      # Configuration
│   ├── database/    # Database implementations
│   ├── redis/       # Cache implementations
│   └── external/    # External API clients
└── presentation/    # Presentation layer
    └── http/       # HTTP routing
```

## 🧪 Testing

### Run Unit Tests
```bash
make test
# OR
go test ./...
```

### Run with Coverage
```bash
make test-coverage
```

### Integration Tests
```bash
make test-integration
```

### Available Make Commands
```bash
make help          # Show all available commands
make build          # Build the application
make run            # Run the application locally
make test           # Run tests
make test-coverage  # Run tests with coverage
make docker-build   # Build Docker image
make docker-up      # Start Docker services
make docker-down    # Stop Docker services
make docker-logs    # View Docker logs
make swagger        # Generate Swagger docs
make lint           # Run linter
make fmt            # Format code
```

## 🐳 Docker Commands

```bash
# Build and start all services
docker compose up -d

# Build image only
docker build -t message-sending-service .

# View logs
docker compose logs -f message-service

# View all service logs
docker compose logs -f

# Restart a specific service
docker compose restart message-service

# Stop services
docker compose down

# Stop and remove volumes (reset database)
docker compose down -v

# Check service status
docker compose ps

# Execute commands in running container
docker compose exec message-service /bin/sh
```

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Logs
The application uses structured logging. Check logs for:
- Message processing status
- Scheduler activity
- API requests/responses
- Error tracking

### Redis Cache
Monitor cached data:
```bash
redis-cli
> KEYS message_sent:*
> GET scheduler_status
```

## 🔧 Development

### Generate Swagger Docs
```bash
make swagger
# OR
swag init -g cmd/server/main.go
```

### Database Migrations
```bash
# Run init script locally
psql -U message_user -d message_db -f scripts/init.sql

# Or with Docker
docker compose exec postgres psql -U message_user -d message_db -f /docker-entrypoint-initdb.d/init.sql
```

### Adding New Features

1. Add domain entities in `internal/domain/entities/`
2. Define repository interfaces in `internal/domain/repositories/`
3. Implement use cases in `internal/application/usecases/`
4. Add HTTP handlers in `internal/application/handlers/`
5. Update routing in `internal/presentation/http/router.go`

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection Error**
   ```bash
   # Check PostgreSQL is running
   brew services list | grep postgresql
   
   # Test connection
   psql -U message_user -d message_db -c "SELECT 1;"
   ```

2. **Redis Connection Error**
   ```bash
   # Check Redis is running
   brew services list | grep redis
   
   # Test connection
   redis-cli ping
   ```

3. **Scheduler Not Starting**
   - Check logs for error messages
   - Verify configuration in `config.env`
   - Ensure database is accessible

4. **External API Failures**
   - Verify `MESSAGE_API_URL` is correct
   - Check if the external service is available
   - Review timeout settings

### Debug Mode
```bash
LOG_LEVEL=debug go run cmd/server/main.go
```

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📞 Support

For questions or issues:
- Create an issue on GitHub
- Check the Swagger documentation
- Review the logs for error details

---

**Built with ❤️ using Go, PostgreSQL, and Redis**