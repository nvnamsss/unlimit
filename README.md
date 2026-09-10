# Unlimit

## Installation Instructions

### Prerequisites

- Go 1.24.0 or higher
- Git

### Steps to Install

1. Clone the repository:
   ```bash
   git clone https://github.com/nvnamsss/unlimit.git
   cd unlimit
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build
   ```

## Dependencies

This project relies on several major dependencies:
- Kafka client (IBM/sarama)
- Redis client (go-redis)
- MongoDB driver
- PostgreSQL and MySQL support (GORM)
- Logging (Zap)

For a complete list of dependencies, see the `go.mod` file.

## Module Structure and Usage

Unlimit is structured as a main module with multiple specialized submodules:

```
unlimit/
├── cache/     # Caching implementations (Redis, etc.)
├── db/        # Database abstractions (MongoDB, PostgreSQL, MySQL)
├── kafka/     # Kafka client integrations
├── log/       # Logging utilities
└── [other submodules...]
```

### Importing the Module

To use Unlimit in your project:

First, export those environment variables:
```
export GOPROXY=direct
export GOPRIVATE=github.com/voidforge-studios/*
```

Then:
```bash
# To get the entire module with all submodules at once:
go get github.com/nvnamsss/unlimit@v0.4.6
```

This single command will make all submodules available for import in your project.

### Using Individual Submodules

Import specific submodules as needed in your code:

```go
import (
    "github.com/nvnamsss/unlimit/cache"
    "github.com/nvnamsss/unlimit/db"
    "github.com/nvnamsss/unlimit/kafka"
    "github.com/nvnamsss/unlimit/log"
)
```

### Example Usage

```go
package main

import (
    "github.com/nvnamsss/unlimit/cache"
    "github.com/nvnamsss/unlimit/log"
)

func main() {
    // Initialize logger
    logger := log.NewZapLogger()
    
    // Initialize Redis cache
    redisCache, err := cache.NewRedisClient("localhost:6379")
    if err != nil {
        logger.Error("Failed to connect to Redis", "error", err)
        return
    }
    
    // Use the cache
    err = redisCache.Set("key", "value", 3600)
    if err != nil {
        logger.Error("Failed to set cache value", "error", err)
    }
}
```

See the documentation in each submodule for more detailed usage instructions.
