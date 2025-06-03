# Build Instructions

This document describes how to build, test, and maintain the GGU (Go Generic Utils) project.

## Prerequisites

- Go 1.18+ (required for generics support)
- make
- git

For linting:
- [golangci-lint](https://golangci-lint.run/usage/install/)

## Using the Makefile

The project includes a Makefile that simplifies common development tasks.

### Available Commands

Run `make help` to see all available commands:

```
$ make help
all                   Run default targets (lint, test)
bench                 Run benchmarks
clean                 Clean up generated files
doc                   Generate and serve Go documentation
examples              Run example applications
fmt                   Format source code
help                  Show help message
lint                  Run linters
test                  Run unit tests
test-coverage         Run tests with coverage
update-copyright      Update copyright headers
update-deps           Update dependencies
vet                   Run go vet
```

### Common Tasks

#### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage
```

#### Code Quality

```bash
# Run linters
make lint

# Format code
make fmt

# Run go vet
make vet
```

#### Running Examples

```bash
# Run example applications
make examples
```

#### Viewing Documentation

```bash
# Generate and serve documentation at http://localhost:6060/pkg/github.com/Humphrey-He/go-generic-utils/
make doc
```

#### Benchmarks

```bash
# Run benchmarks
make bench
```

## CI/CD

The project uses GitHub Actions for continuous integration. The workflow is defined in `.github/workflows/ci.yml` and includes:

- Testing on multiple Go versions (1.18, 1.19, 1.20, 1.21)
- Code linting with golangci-lint
- Code coverage reporting

## Manual Build

If you don't want to use the Makefile, you can use standard Go commands:

```bash
# Run tests
go test ./...

# Run tests with race detection
go test -race ./...

# Format code
go fmt ./...

# Run vet
go vet ./...
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details. 