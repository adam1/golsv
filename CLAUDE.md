# golsv Coding Guidelines

## Build Commands
- `make build`: Build main package and CLI tools
- `make test`: Run all tests
- `go test ./...`: Run all tests with standard output
- `go test -v ./path/to/package`: Run tests in specific package with verbose output
- `go test -run TestName`: Run a single test function
- `make coverage`: Show test coverage summary
- `make watch`: Watch for file changes and rebuild/test

## Code Style
- Package: Use single package `golsv` for all main code
- Imports: Group in one block, standard library first, then third-party
- Types: Use type parameters for generics; prefer interfaces for abstraction
- Naming: PascalCase for exported identifiers, camelCase for unexported
- Error handling: Panic for programmer errors, return errors for expected failures
- Testing: Table-driven tests with clear input/output descriptions
- Math notation: Use standard mathematical notation in comments
- Documentation: All exported functions and types should have comments