# Backend Development Guidelines

## Introduction
This document serves as a comprehensive guide for developing backend services using the Remembrance project structure as a reference. Following these guidelines will ensure consistency, maintainability, and scalability across all backend services.

## Table of Contents
1. [Project Structure](#project-structure)
2. [Architecture Overview](#architecture-overview)
3. [Development Workflow](#development-workflow)
4. [API Design](#api-design)
5. [Error Handling](#error-handling)
6. [Database Operations](#database-operations)
7. [Logging](#logging)
8. [Configuration Management](#configuration-management)
9. [Authentication & Authorization](#authentication--authorization)
10. [Messaging with Kafka](#messaging-with-kafka)
11. [Testing](#testing)
12. [Deployment](#deployment)
13. [Kubernetes Configuration](#kubernetes-configuration)

## Project Structure

The backend service follows a clean architecture with clear separation of concerns:

```
src/
├── adapters/          # External service interfaces (DB, cache, etc.)
├── cmd/               # Entry points for different executables
│   ├── main.go        # API server entry point
│   ├── consumer/      # Consumer service entry point
│   └── docs/          # API documentation
├── configs/           # Configuration management
├── controllers/       # HTTP request handlers
├── dtos/              # Data transfer objects
├── errors/            # Error definitions
├── handlers/          # Message handlers (e.g., Kafka)
├── logger/            # Logging mechanisms
├── middlewares/       # HTTP middleware components
├── migrations/        # Database migrations
├── models/            # Domain models
├── pkg/               # Shared packages
├── repositories/      # Data access layer
├── services/          # Business logic
└── utils/             # Utility functions
```

## Architecture Overview

Our backend services follow a layered architecture:

1. **Controllers Layer** - Handles HTTP requests and responses
2. **Service Layer** - Implements business logic
3. **Repository Layer** - Manages data access
4. **Adapters Layer** - Interfaces with external services

The flow of data is:
- HTTP Request → Controller → Service → Repository → Database
- Kafka Message → Handler → Service → Repository → Database

## Development Workflow

### Setting up a New Service

1. Create the necessary directories following the project structure
2. Define your models in the `models` package
3. Create DTOs in the `dtos` package
4. Implement repositories in the `repositories` package
5. Implement services in the `services` package
6. Create controllers in the `controllers` package
7. Set up routes in `cmd/main.go`

### Adding a New Feature

1. Define models and DTOs
2. Implement repository methods
3. Add service logic
4. Create controller endpoints
5. Register routes
6. Add migration if needed
7. Update configuration

## API Design

### Endpoint Structure

- Use RESTful design principles 
- Group related endpoints under a common prefix
- Use versioning in the URL path (e.g., `/v1/event`)

### Request/Response Format

- Use JSON for request and response bodies
- Use DTOs for data validation and transformation
- Follow consistent naming conventions

Example controller method:

```go
// GET /v1/event/:id
func (h *EventController) Get(c *gin.Context) {
    var id int64
    var err error
    
    if id, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
        h.HandleError(c, errors.New(errors.ErrInvalidRequest, err.Error()))
        return
    }
    
    res, err := h.eventService.Get(c.Request.Context(), id)
    if err != nil {
        h.HandleError(c, err)
        return
    }
    
    h.JSON(c, res)
}
```

## Error Handling

- Define application-specific errors in the `errors` package
- Use error codes and consistent error responses
- Include contextual information in errors
- Log errors with appropriate severity levels

Example:

```go
if event == nil {
    logger.Context(ctx).Errorf("event %v not found", id)
    return nil, errors.New(errors.ErrNotFound)
}
```

## Database Operations

### Migrations

- Use SQL migrations in the `migrations` folder
- Name migrations with a timestamp prefix and descriptive name
- Add comments for up/down operations
- Test migrations before applying them

### Repository Pattern

- Each domain entity should have its own repository
- Define repository interfaces in the `repositories` package
- Implement CRUD operations and custom queries

Example repository interface:

```go
type EventRepository interface {
    Create(ctx context.Context, event *models.Event) error
    Get(ctx context.Context, id int64) (*models.Event, error)
    GetByCode(ctx context.Context, code string) (*models.Event, error)
    Search(ctx context.Context, req *dtos.SearchEventRequest) ([]*models.Event, int64, error)
    Update(ctx context.Context, event *models.Event) error
}
```

## Logging

- Use structured logging via the `logger` package
- Include context in log messages
- Use appropriate log levels (Info, Error, Debug)
- Log the beginning and end of important operations
- Include request IDs for traceability

Example:
```go
logger.Context(ctx).Infof("Starting operation X with params: %v", params)
// operation code
logger.Context(ctx).Infof("Completed operation X successfully")
```

## Configuration Management

- Use environment variables for configuration
- Define configuration structures in the `configs` package
- Provide sensible defaults
- Validate configuration on startup

Example:
```go
type App struct {
    Host string `default:"0.0.0.0" envconfig:"HOST"`
    Port int    `default:"8080" envconfig:"PORT"`
    // ...
}
```

## Authentication & Authorization

- Use JWT for authentication
- Implement middleware for authorization
- Apply authorization middleware to route groups
- Include user information in the context

Example middleware usage:
```go
// Protect routes
event := v1.Group("/event")
{
    event.Use(middlewares.VerifyJWT())
    event.GET("", eventController.List)
    // ...
}
```

## Messaging with Kafka

- Define message structures in the `dtos` package
- Create handlers in the `handlers` package
- Use producer clients from the `pkg/queue` package
- Handle failures and retries appropriately

Example sending a message:
```go
utils.SendMessage(ctx, s.kafkaProducer, configs.AppConfig.Topics.Notification, &dtos.KafkaMessage[dtos.NotificationPayload]{
    ID:        utils.UUID(),
    Event:     models.EventNotificationCreate,
    Timestamp: time.Now().Unix(),
    Payload: dtos.NotificationPayload{
        Event:  models.OnEventCreated,
        UserID: req.CreatorID,
        Url:    fmt.Sprintf("/remembrance/%v", event.ID),
        Data:   data,
    },
})
```

## Testing

- Write unit tests for services and repositories
- Use mocks for external dependencies
- Create integration tests for critical paths
- Test error scenarios
- Aim for high test coverage

## Deployment

### Docker

- Use multi-stage builds for smaller images
- Include only necessary files
- Set proper environment variables
- Use non-root users for security

### CI/CD

- Use automated builds and tests
- Deploy automatically to development environments
- Require manual approval for production deployments
- Include database migration steps

## Kubernetes Configuration

- Define resource limits and requests
- Configure health checks
- Set up proper scaling rules
- Use Kubernetes secrets for sensitive data
- Configure service and ingress resources appropriately

Example Kubernetes deployment structure:
```
k8s/
├── dev/
│   ├── config.dev.yml
│   ├── deploy.dev.migration.sh
│   ├── deploy.dev.sh
│   ├── main.consumer.yml
│   ├── main.migration.yml
│   ├── main.mobile.yml
│   └── main.ops.yml
```

## Best Practices

1. Follow Go coding conventions
2. Use context for request-scoped operations
3. Handle errors appropriately at each layer
4. Use dependency injection for better testability
5. Keep services stateless
6. Cache where appropriate
7. Use transactions for multi-step database operations
8. Include proper documentation with Swagger annotations
9. Monitor and instrument code for observability
10. Maintain consistent logging practices

---

This guideline document is intended to evolve with the project. Feel free to suggest improvements or clarifications.

<!-- Generated by Copilot -->