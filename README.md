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

## 🛠️ Installation

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/message-sending-service.git
   cd message-sending-service
   ```

2. **Configure environment variables**
   ```bash
   cp config.env.example config.env
   # Edit config.env with your settings
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Initialize the database**
   ```bash
   docker-compose exec postgres psql -U insider -d insider_messaging -f /docker-entrypoint-initdb.d/init.sql
   ```

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
   createuser -s insider
   createdb -O insider insider_messaging
   
   # Run initialization script
   psql -U insider -d insider_messaging -f scripts/init.sql
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
DB_USER=insider
DB_PASSWORD=insider123
DB_NAME=insider_messaging
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

### Create a Message
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, this is a test message!",
    "phone_number": "+1234567890"
  }'
```

### Start Automatic Sending
```bash
curl -X POST http://localhost:8080/api/v1/scheduler/start
```

### Get Scheduler Status
```bash
curl http://localhost:8080/api/v1/scheduler/status
```

### Get Sent Messages
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

## 🐳 Docker Commands

```bash
# Build image
docker build -t message-sending-service .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f message-service

# Stop services
docker-compose down

# Reset database
docker-compose down -v
docker-compose up -d
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
# Run init script
psql -U insider -d insider_messaging -f scripts/init.sql
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
   psql -U insider -d insider_messaging -c "SELECT 1;"
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