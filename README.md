# Cosmic Go 🚀

A sophisticated Go implementation of the architecture patterns and DDD concepts from [Cosmic Python](https://www.cosmicpython.com/), demonstrating enterprise-level software design and best practices.

## 🎯 Project Overview

This project showcases advanced software engineering practices including:

- **Domain-Driven Design (DDD)**: Implementing a rich domain model with clear boundaries and business rules
- **Clean Architecture**: Separation of concerns through layered architecture
- **SOLID Principles**: Especially Dependency Inversion through interfaces and ports/adapters pattern
- **Event-Driven Architecture**: Using Redis for message publishing/subscribing
- **Advanced Testing Strategies**: Unit, integration, and E2E tests with test containers
- **Concurrency Handling**: Thread-safe operations with proper locking mechanisms
- **Infrastructure as Code**: Docker Compose for local development

## 🏗️ Architecture

The project follows a clean, hexagonal architecture:

```plaintext
├── cmd/            # Application entrypoints
├── internal/       # Private application code
│   ├── domain/     # Domain model and business rules
│   ├── application/# Application services and use cases
│   ├── infra/      # Infrastructure implementations
│   └── handler/    # HTTP handlers and API endpoints
├── pkg/            # Public shared utilities
└── test/           # Test suites and helpers
```

## 🛠️ Technical Stack

- **Go 1.23+**: Modern Go features and idioms
- **GORM**: Sophisticated ORM with PostgreSQL driver
- **Redis**: Message broker for event handling
- **Docker**: Containerization and local development
- **Testcontainers**: Integration testing with real dependencies
- **PostgreSQL**: Primary data store
- **PgAdmin**: Database management UI

## 🚀 Getting Started

1. **Prerequisites**
   - Go 1.23+
   - Docker and Docker Compose
   - Make (optional)

2. **Setup Local Environment**
   ```bash
   # Start infrastructure dependencies
   make run-dependencies
   
   # Run the application
   make run
   ```

3. **Run Tests**
   ```bash
   # Unit tests
   make unit-test
   
   # Integration tests
   make integration-test
   
   # End-to-end tests
   make e2e-test
   
   # Generate coverage report
   make coverage
   ```

## 🧪 Testing Strategy

The project implements a comprehensive testing pyramid:

- **Unit Tests**: Fast tests focusing on domain logic and business rules
- **Integration Tests**: Testing repository patterns and database interactions
- **E2E Tests**: Full system tests using test containers for real dependencies

## 📚 Key Features

- **Rich Domain Model**: Sophisticated business logic implementation
- **CQRS Pattern**: Separate read and write operations
- **Event Sourcing**: Track state changes through domain events
- **Optimistic Concurrency**: Handle concurrent operations safely
- **Repository Pattern**: Abstract data persistence
- **Unit of Work**: Maintain data consistency
- **Message Publishing**: Async communication between components

## 🔍 Code Quality

- Strict linting rules with golangci-lint
- Comprehensive test coverage
- Documentation following Go best practices
- Clean code principles and SOLID design

## 📖 Learning Resources

This project demonstrates practical implementations of concepts from:
- [Cosmic Python](https://www.cosmicpython.com/)
- Domain-Driven Design principles
- Clean Architecture patterns
- Go best practices

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
```