# Introduction
This repository is 95% AI, 5% human. It was created for the purpose of testing a new approach to backend development with the assistance of LLMs. It follows a set of guidelines composed from a real project by itself (`backend_development_guidelines.md`), but is primarily an experimental implementation to evaluate the effectiveness of LLM-assisted development giving.

From my perspective, the LLM can replicate about 85% of my coding style. The discrepancies mainly stem from the frameworks and libraries used in the repository. This could be improved by introducing another context file named software_requirements.md, which would declare the necessary software dependencies. These additional requirements would guide the LLM in making more accurate implementation choices.

In summary, backend_development_guidelines.md defines the architecture for the LLM, while software_requirements.md outlines the frameworks and libraries to be used.


# Chat Service

A Go backend service that handles user chat interactions with LLM models. This service is responsible for managing chat sessions, storing message history, and communicating with an external LLM service.

## Architecture

The Chat service follows a clean architecture pattern with the following components:

- **Models**: Domain entities representing chats and messages
- **DTOs**: Data Transfer Objects for API requests and responses
- **Repositories**: Data access layer for database operations
- **Services**: Business logic layer
- **Controllers**: HTTP request handlers
- **Adapters**: External service integrations (database, LLM)
- **Middlewares**: HTTP middleware components (auth, logging, etc.)

## Features

- Create and manage chat sessions
- Store and retrieve chat history
- Process user messages through integration with LLM vendor service
- Authentication via JWT tokens
- Event-based architecture with Kafka integration
- Structured logging and request tracing

## API Endpoints

### Chat Management

- `POST /api/v1/chats` - Create a new chat
- `GET /api/v1/chats` - List all chats for a user
- `GET /api/v1/chats/search` - Search chats by title
- `GET /api/v1/chats/:id` - Get a specific chat
- `PUT /api/v1/chats/:id` - Update a chat
- `DELETE /api/v1/chats/:id` - Delete a chat

### Message Management

- `POST /api/v1/messages?chatId=<id>` - Send a message to a chat
- `GET /api/v1/messages?chatId=<id>` - List messages for a chat
- `GET /api/v1/messages/:id` - Get a specific message
- `PUT /api/v1/messages/:id` - Update a message
- `DELETE /api/v1/messages/:id` - Delete a message

## Setup

### Prerequisites

- Go 1.20+
- PostgreSQL 13+
- Kafka (optional, mock implementation included)

### Environment Variables

The service can be configured using environment variables or a YAML config file. See `config.yaml` for available options.

Required environment variables:

```
DB_HOST - Database host
DB_USER - Database user
DB_PASSWORD - Database password
DB_NAME - Database name
LLM_BASE_URL - URL of the LLM vendor service
LLM_API_KEY - API key for the LLM vendor service
JWT_SECRET - Secret key for JWT token signing
```

### Running the Service

#### Local Development

1. Clone the repository
2. Set up the PostgreSQL database
3. Copy `config.yaml` to your preferred location and update values
4. Run the service:

```bash
go run src/cmd/main/main.go -config=config.yaml
```

#### Using Docker

```bash
docker build -t chat-service .
docker run -p 8080:8080 --env-file .env chat-service
```

## Development

### Database Migrations

The service uses `golang-migrate` for database migrations. Migrations are automatically applied when the service starts.

To create new migrations:

```bash
migrate create -ext sql -dir src/migrations -seq <migration_name>
```

### Adding New Features

To add new features:

1. Update models and DTOs as necessary
2. Implement repository methods if needed
3. Add business logic to the service layer
4. Create or update controller endpoints
5. Test the changes

## Testing

Run the tests with:

```bash
go test ./...
```

## Integration with Other Services

This service integrates with:

- **LLM Vendor Service**: A Python service that handles LLM interactions
- **User Service**: Provides authentication and user information (via JWT)
- **Analytics Services**: Consumes events from Kafka topics for analytics

## Project Background

This project demonstrates how large language models can assist in implementing backend services when given clear architectural guidelines. The implementation process involved:

1. Providing the LLM with the `backend_development_guidelines.md` document
2. Defining the core requirements for a chat service
3. Allowing the LLM to generate the necessary components following the guidelines
4. Human review and iterative improvement of the generated code

This approach shows promising results for accelerating development while maintaining architectural consistency and best practices. The code structure, patterns, and quality closely follow what would be expected in a production environment, while benefiting from the speed and knowledge breadth of LLM assistance.
