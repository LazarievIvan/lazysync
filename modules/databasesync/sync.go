package databasesync

import "fmt"

type SyncDatabase struct {
	config *Config
}

func (db *DB) Sync() {
	fmt.Println("Sync started")
}
