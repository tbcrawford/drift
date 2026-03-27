# Default: show available recipes
default:
    @just --list

# Run all tests
test:
    go test ./...

# Run tests with race detector
test-race:
    go test -race ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Build CLI binary
build:
    go build ./cmd/drift/...

# Install CLI to $GOBIN or `go env GOPATH`/bin; add that directory to PATH to run `drift`
install:
    go install ./cmd/drift

# Run linter
lint:
    golangci-lint run ./...

# Run go vet
vet:
    go vet ./...

# Tidy dependencies
tidy:
    go mod tidy

# Run property-based tests with verbose output
test-property:
    go test -v -run TestProperty ./...

# Fuzz the Myers algorithm (60s budget)
fuzz:
    go test -fuzz=FuzzMyers -fuzztime=60s ./internal/algo/myers/...
