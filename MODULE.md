# Unlimit Module Management Guide

This document provides guidelines for both users and contributors on how to work with the Unlimit module system.

## Module Structure

Unlimit follows a standard Go module structure with submodules:

```
unlimit/
├── go.mod                  # Main module definition
├── go.sum                  # Dependency checksums
├── main.go                 # Package documentation and shared functionality
├── cache/                  # Caching submodule
│   ├── redis.go
│   ├── memory.go
│   └── interface.go
├── db/                     # Database submodule
│   ├── mongo/              # MongoDB implementation
│   ├── sql/                # SQL databases implementation
│   └── interface.go
├── kafka/                  # Kafka integration
├── log/                    # Logging utilities
└── [other submodules]
```

## For Users

### Importing the Module

To use Unlimit in your project:

```bash
go get github.com/voidforge-studios/unlimit
```

This will download all submodules, and you can import only what you need:

```go
import (
    "github.com/voidforge-studios/unlimit/cache"
    "github.com/voidforge-studios/unlimit/db/mongo"
)
```

### Minimal Imports

If you only need specific functionality, you can import just those submodules:

```go
import "github.com/voidforge-studios/unlimit/log"
```

This will only pull in the logging functionality without the database or other components.

## For Contributors

### Adding New Submodules

1. Create a new directory for your submodule
2. Ensure it has a clear package declaration
3. Create interface.go to define the public API
4. Implement the functionality in separate files
5. Add unit tests in a `_test.go` file

### Versioning

We follow semantic versioning. When making changes:

- Bug fixes: increment the patch version (1.0.0 → 1.0.1)
- New features (backward compatible): increment the minor version (1.0.0 → 1.1.0)
- Breaking changes: increment the major version (1.0.0 → 2.0.0)

### Publishing Updates

1. Tag releases using Git tags that follow semantic versioning
2. Push tags to make the new version available to users
3. Update documentation to reflect changes

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Best Practices for Module Development

1. Keep dependencies minimal for each submodule
2. Use interfaces for flexibility and testability
3. Provide comprehensive documentation
4. Include examples in each submodule
5. Follow Go's idiomatic patterns
