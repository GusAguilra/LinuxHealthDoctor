# Contributing to Linux Health Doctor

## Getting Started

1. Fork the repository
2. Clone your fork
3. Run `make build` to verify the build
4. Run `make test` to run tests

## Development Workflow

1. Create a feature branch
2. Make your changes
3. Run `make lint && make test` to verify
4. Commit with conventional commit messages
5. Push and open a pull request

## Code Style

- **No comments in code** — Write self-documenting code with clear naming
- **No emojis** — In code, commits, or documentation unless user-facing
- Follow Go standard formatting (`go fmt`)
- Run `make lint` before committing
- Write tests for all new functionality

## Commit Messages

Follow conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Testing
- `refactor:` Code refactoring
- `chore:` Maintenance

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure CI passes
4. Request review from maintainers

## Adding New Health Checks

1. Create a new package under `internal/checks/<name>/`
2. Implement the `plugin.Checker` interface
3. Register with `plugin.Register()` in `init()`
4. Add tests
5. Add knowledge base rules in `internal/knowledge/builtin/`

## Adding New Distribution Support

1. Create a new file in `internal/distro/`
2. Implement the `distro.Distro` interface
3. Add detection logic in `detection.go`

## Code of Conduct

Be respectful, inclusive, and professional. We welcome contributions from everyone.
