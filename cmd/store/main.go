package main

import (
	"fmt"

	"github.com/George-Yanev/go-fun/internal/store"
)

func main() {
	store := store.New(int64(100))
	store.Set("test", 64)
	item, exists := store.Get("test")
	fmt.Printf("exists: %t, value: %+v\n", exists, item)
}
