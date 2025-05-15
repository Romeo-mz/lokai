# Contributing to lokai

Thank you for considering contributing to lokai! This guide will help you get started.

## Development Setup

```bash
# Clone the repo
git clone https://github.com/romeo-mz/lokai.git
cd lokai

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linters
make lint
```

### Prerequisites

- **Go 1.24+**
- **Ollama** installed and running (for integration testing)
- **golangci-lint** (optional, for linting)

## Project Structure

```
cmd/lokai/          → CLI entry point
internal/
  benchmark/        → Model benchmarking engine
  hardware/         → Hardware detection (CPU, RAM, GPU)
  models/           → Model catalog, recommendation engine, performance estimation
  ollama/           → Ollama client wrapper
  ui/               → Terminal UI (charmbracelet)
```

## Making Changes

1. **Fork** the repository
2. Create a **feature branch**: `git checkout -b feat/my-feature`
3. Make your changes
4. **Test**: `make test`
5. **Lint**: `make lint`
6. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat: add new model to catalog`
   - `fix: correct VRAM detection on AMD`
   - `docs: update README`
7. **Push** and open a Pull Request

## Adding Models

To add a new model to the catalog, edit `internal/models/database.go`:

```go
{
    Name: "Model Name", OllamaTag: "model:tag", Family: "family",
    ParameterSize: "7B", ParameterCount: 7.0, QuantLevel: "Q4_K_M",
    DiskSizeGB: 4.0, EstimatedVRAMGB: 6.0,
    UseCases: []UseCase{UseCaseChat}, Quality: 55,
    Description: "Short description",
},
```

Make sure to:
- Set accurate `EstimatedVRAMGB` (test with `ollama show <model>`)
- Assign a realistic `Quality` score (1-100)
- Place the entry in the correct category section
- Keep entries ordered by quality within their category

## Reporting Issues

- Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md)
- Include your hardware specs (`lokai --scan-only --json`)
- Include the full error output

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and small
- Add comments only when the logic isn't self-evident
- No external dependencies without discussion

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
