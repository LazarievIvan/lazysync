package database

import "fmt"

type SyncDatabase struct {
	config *Config
}

func Sync() {
	fmt.Println("Sync started")
}
