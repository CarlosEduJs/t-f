# Contributing to t-f

Thank you for your interest in contributing to t-f. This project aims to be a reliable and predictable tool for converting CSS variables to design tokens.

## Getting Started

1. Clone the repository.
2. Ensure you have Go 1.26 or later installed.
3. Run `make install-hooks` to set up the local development environment with pre-commit checks.

## Development Workflow

We use a Makefile to manage common tasks. Before submitting any changes, please ensure that your code passes all checks:

- Run `make lint` to check for style and quality issues.
- Run `make test` to ensure all tests pass.
- Run `make build` to verify that the project compiles correctly.

## Coding Standards

- Follow idiomatic Go patterns.
- Ensure all exported types and functions have descriptive doc comments.
- Use explicit error handling and wrap errors where it adds useful context (e.g., `fmt.Errorf("...: %w", err)`).
- Avoid unnecessary external dependencies to keep the project lightweight and maintainable.

## Submitting Changes

1. Create a new branch for your feature or bug fix.
2. Commit your changes with clear, descriptive commit messages.
3. Ensure the pre-commit hooks (lint and tests) pass before pushing.
4. Open a Pull Request with a detailed description of your changes.

## Testing

When adding new features, please include corresponding tests in the relevant `_test.go` files. We aim for high test coverage for the core parsing and generation logic.
