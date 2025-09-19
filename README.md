![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)
[![Go Reference](https://pkg.go.dev/badge/github.com/Serajian/query-builder-GO.svg)](https://pkg.go.dev/github.com/Serajian/query-builder-GO)
# 🗄️ cache-db

A lightweight **in-memory key-value store** for Go with:
- Generic key & value types
- Per-key TTL (expiration)
- Thread-safe access (RWMutex)
- Persistence with `gob` (atomic writes with temp + rename)
- Simple file management (`Load`, `Persist`, `DeleteFile`)

---

### 🚀 Installation

```bash
go get github.com/Serajian/cache-db@latest
```

---

### 📦 Usage
Define Database
```go
package main

import (
	"fmt"
	"time"
	"github.com/Serajian/cache-db.git/database"
)

func main() {
	// Create a DB with string keys and string values.
	// Default TTL = 2 seconds, data persisted under ./data directory.
	db := database.NewDatabase[string, string](2*time.Second, "./data")

	// Set a key with default TTL
	db.Set("foo", "bar")

	// Persist to disk
	if err := db.Persist("test.gob"); err != nil {
		panic(err)
	}

	// Load into a fresh DB
	db2 := database.NewDatabase[string, string](0, "./data")
	if err := db2.Load("test.gob"); err != nil {
		panic(err)
	}

	// Retrieve
	if v, ok := db2.Get("foo"); ok {
		fmt.Println("Loaded value:", v)
	}

	// TTL expiry check
	time.Sleep(3*time.Second)
	if _, ok := db2.Get("foo"); !ok {
		fmt.Println("foo expired as expected")
	}

	// Delete file
	if err := db2.DeleteFile("test.gob"); err != nil {
		panic(err)
	}
}

```

---

### 📂 Features

#### Set / Get / Delete / Clear

##### -TTL support: Expiration per-key or default for DB

##### -CleanExpired: Manually purge expired entries

#### Persistence:

#### -Persist(filename) → Save DB atomically to disk

#### -Load(filename) → Load DB state from disk

#### -DeleteFile(filename) → Remove persisted file safely

#### -Thread-safe with sync.RWMutex

---

### 🛠 Model Types

#### Entry[V] → Wraps value with ExpiresAt time.Time

#### Persisted[K,V] → On-disk format (with version & defaultTTL)

---
## 🤝 Contributing

Feel free to contribute by opening pull requests to help improve and extend this project.  
Your contributions are always welcome!
---

## License

[MIT License](LICENSE.txt)

