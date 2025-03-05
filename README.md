# go-playground
This repository is a sandbox for exploring and experimenting with key Go programming concepts such as concurrency, error handling, interfaces, and more. It contains a collection of small, self-contained projects that tries to use Go in idiomatic way.

# Projects
Each project is organized in its own subfolder under cmd/, where the main.go file serves as the entry point. Supporting logic and reusable functions are typically located in the internal/ folder to encapsulate implementation details and promote modularity.
Currently, the repository includes the following projects:

## Billion Row Challenge (1BRC)
**Description:** Solution to the 1 Billion Row Challenge
**Location:** cmd/1brc/

## Key/Value Store with TTL
**Description:** A simple in-memory key/value store with support for time-to-live (TTL) expiration of entries.
**Location:** cmd/store/

## URL Shortener
**Description:** A basic URL shortening service that generates short aliases for long URLs and redirects to the original links.
**Location:** cmd/urlshortener/

# Project Structure
```
go-playground/
├── cmd/              # Entry points for each project
│   ├── 1brc/         # 1 Billion Row Challenge solution
│   ├── store/        # Key/Value store with TTL
│   └── urlshortener/ # URL shortener
├── internal/         # Shared logic and utilities
└── README.md         # This file
```

# Running the Projects
1. Clone the repository:

```bash
git clone https://github.com/<your-username>/go-playground.git
cd go-playground
```

2. Run it

```bash
go run cmd/1brc/main.go
```
