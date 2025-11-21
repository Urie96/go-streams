# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library called `streams` that provides **lazy evaluation** utility functions for streaming data processing without goroutines or channels, supporting partial stream consumption. The library is designed to solve common pain points with channel-based streaming by avoiding goroutine leaks, mandatory full consumption, and complex concurrent scheduling.

## Core Architecture

### Stream Interface
The foundation of the library is the `Stream[T any]` interface with a single method:
```go
type Stream[T any] interface {
    Recv() (T, error)
}
```

### Key Components

- **BaseStream** (`base.go`): Core stream implementation with `Consume()` method for processing stream items
- **Factory Functions**: Create streams from slices, channels, functions, errors (`create.go`)
- **Transform Functions**: Map, filter, and transform streams (`higher_order.go`, `map.go` equivalents)
- **Combining Functions**: Concat, fork, merge streams (`concat.go`, `fork.go`, `merge.go`)
- **Label Processing**: Token-based stream labeling and demultiplexing (`label.go`, `special_token.go`)
- **String Utilities**: Specialized string stream processing (`string_reader.go`, `remove_token.go`)
- **Safety Utilities**: Concurrent-safe streams and logging (`safe.go`, `with_log.go`)

### Design Philosophy

- **Lazy Evaluation**: Operations are only executed when the stream is consumed
- **No Goroutine Leaks**: Avoids the common pattern of requiring complete channel consumption
- **Type Safety**: Uses Go 1.22+ generics for type-safe streaming operations
- **Composable**: Streams can be chained and combined in various ways

## Common Development Commands

### Testing
```bash
# Run all tests with coverage
make test
# Or directly:
go test -v ./... -cover

# Run specific test file
go test -v ./streams -run TestSpecificFunction

# Run integration tests
go test -v ./streams -run Integration
```

### Building
```bash
# Build the module (no build output expected for library)
go build ./...
```

### Module Management
```bash
# Tidy dependencies
go mod tidy

# Download dependencies
go mod download
```

## Key Patterns

### Creating Streams
```go
// From slice
sliceStream := streams.FromSlice([]string{"hello", "world"})

// From channel
chStream := streams.FromChan(make(chan string))

// From function
funcStream := streams.FromFunc(func() (string, error) {
    // Return next value or io.EOF
})

// From error (creates a stream that immediately returns the error)
errStream := streams.FromErr[string](someError)
```

### Stream Processing Pipeline
```go
// Transform stream
mapped := streams.Map(input, func(item string) string {
    return strings.ToUpper(item)
})

// Filter content
filtered := streams.SkipFunc(input, func(item string) bool {
    return strings.Contains(item, "skip")
})

// Combine streams
combined := streams.Concat(stream1, stream2)
```

### Consuming Streams
```go
// Process each item
streams.Consume(stream, func(item string) error {
    // Handle item
    return nil
})

// Collect all items
items := streams.Collect[string](stream)

// Convert to channel (with goroutine - ensure full consumption)
ch := streams.ToChan(stream)
```

## Testing Strategy

The codebase includes comprehensive test coverage:
- Unit tests for each stream operation (`*_test.go` files)
- Integration tests (`integration_test.go`) testing combined operations
- Error handling validation
- Edge case testing (empty streams, nil inputs, etc.)

## Important Notes

- The library uses Go 1.22+ generics extensively
- All operations are lazy - no computation happens until stream consumption
- `ToChan()` creates a goroutine and requires full channel consumption to avoid leaks
- Error handling varies by operation - some ignore errors, others propagate them
- The `BaseStream` wrapper provides nil-safe stream operations