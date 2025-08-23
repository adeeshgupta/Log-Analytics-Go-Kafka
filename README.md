# Log Analytics & Monitoring System

A real-time log processing and monitoring system built with Go, Apache Kafka, and MySQL with integrated alerting capabilities.

## Architecture Overview

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐    ┌──────────────┐
│   Application   │───▶│    Kafka     │───▶│   Go Services   │───▶│    MySQL     │
│     Logs        │    │   (Stream)   │    │  (Processing)   │    │   (Storage)  │
└─────────────────┘    └──────────────┘    └─────────────────┘    └──────────────┘
                                │                       │
                                ▼                       ▼
                       ┌──────────────┐    ┌─────────────────┐
                       │   Real-time  │    │   Alert System  │
                       │   Dashboard  │    │   (Database)    │
                       └──────────────┘    └─────────────────┘
```

## Architecture Approach

The system uses a **hybrid architecture** for optimal development experience:

### Docker Infrastructure
- **Kafka** - Message streaming and event processing
- **Zookeeper** - Kafka coordination and metadata
- **MySQL** - Data storage for logs, alerts, and metrics
- **Kafka UI** - Web interface for Kafka management and monitoring

### Go Services
- **Log Collector** - Generates and sends logs to Kafka
- **Log Processor** - Consumes logs from Kafka and stores in MySQL
- **API Server** - Provides REST API and dashboard

## Features

- **Real-time Log Ingestion**: Collect logs from multiple applications via Kafka
- **Log Processing**: Parse, filter, and enrich log data using Go services
- **Pattern Detection**: Identify error patterns, performance issues, and anomalies
- **Real-time Dashboard**: Monitor system health and log metrics
- **Log Search**: Full-text search across processed logs
- **Performance Metrics**: Track response times, error rates, and throughput
- **Alert System**: Configurable alert rules with real-time monitoring
- **Alert Management**: Create, acknowledge, and resolve alerts
- **Alert Statistics**: Comprehensive alert analytics and reporting

## Project Structure

```
├── cmd/                    # Application entry points
│   ├── log-collector/     # Kafka producer for log ingestion
│   ├── log-processor/     # Kafka consumer for log processing
│   └── api-server/        # REST API and dashboard
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── constants/        # Application constants
│   ├── database/         # MySQL operations
│   │   ├── alerts/       # Alert repository
│   │   └── logs/         # Log repository
│   ├── handlers/         # HTTP handlers
│   ├── kafka/            # Kafka producer/consumer
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   └── services/         # Business logic services
├── scripts/              # Database migrations
│   └── migrations/       # SQL migration files
├── go.mod               # Go module file
├── Makefile             # Build and run commands
└── README.md            # This file
```

## Quick Start

### Development Setup

1. **Configure environment**:
   ```bash
   cp env.example .env
   ```

2. **Start Docker infrastructure**:
   ```bash
   make docker-up
   ```

3. **Setup database and run migrations**:
   ```bash
   make migrate
   ```

4. **Build and run services**:
   ```bash
   # Build all binaries
   make build
   
   # Terminal 1: Log processor
   make run-processor
   
   # Terminal 2: API server
   make run-api
   
   # Terminal 3: Log collector
   make run-collector
   ```

5. **Access services**:
   - Dashboard: http://localhost:8080
   - Kafka UI: http://localhost:8081 (Kafka management interface)


## API Endpoints

### Log Endpoints
- `GET /api/logs` - Search logs with filters
- `GET /api/logs/trace/:traceID` - Get logs by trace ID
- `GET /api/metrics` - Get system metrics and statistics
- `GET /api/health` - Health check endpoint

### Alert Endpoints
- `GET /api/alerts` - Get alerts with filters
- `GET /api/alerts/stats` - Get alert statistics
- `GET /api/alerts/active` - Get active alerts
- `GET /api/alerts/:id` - Get alert by ID
- `PUT /api/alerts/:id/resolve` - Resolve an alert
- `PUT /api/alerts/:id/acknowledge` - Acknowledge an alert

### Alert Rule Endpoints
- `POST /api/alert-rules` - Create a new alert rule
- `GET /api/alert-rules` - Get all alert rules
- `GET /api/alert-rules/:id` - Get alert rule by ID
- `PUT /api/alert-rules/:id` - Update an alert rule
- `DELETE /api/alert-rules/:id` - Delete an alert rule

## Alert System

### Alert Rules
Alert rules define conditions that trigger alerts when met. Each rule includes:
- **Name**: Human-readable rule name
- **Description**: Rule description
- **Condition**: SQL condition to evaluate (e.g., error rate, response time)
- **Threshold**: Value that triggers the alert
- **Time Window**: Time period to evaluate (in minutes)
- **Severity**: Alert severity level (low, medium, high, critical)
- **Enabled**: Whether the rule is active

## Available Commands

```bash
make help          # Show all available commands
make build         # Build all Go binaries
make migrate       # Run database migrations
make migrate-status # Show migration status
make run-processor # Run log processor service
make run-api       # Run API server and dashboard
make run-collector # Run log collector service
make clean         # Clean build artifacts
```

## Documentation
- **`Makefile`** - Available build and run commands
- **`docker-compose.yml`** - Docker infrastructure configuration

## Migration System

The project uses a migration-based approach for database setup:

### Migration Files
- `000_create_database.sql` - Creates the database and switches to it
- `001_create_migrations_table.sql` - Creates the migrations table for tracking
- `002_initial_schema.sql` - Creates all tables (logs, alert_rules, alerts)
- `003_sample_alert_rules.sql` - Inserts sample alert rules
- `004_sample_data.sql` - Inserts sample log data
