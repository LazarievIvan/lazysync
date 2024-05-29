package databasesync

const ID = "databasesync"

type DB struct {
	id             string
	DatabaseConfig *Config
}

func (db *DB) SetupModule() {
	db.DatabaseConfig = &Config{
		Host:     "localhost",
		Port:     "80",
		Username: "admin",
		Password: "admin",
		Database: "main",
	}
}

func (db *DB) GetConfiguration() map[string]string {
	return db.DatabaseConfig.GetConfigMap()
}

func (db *DB) GetId() string {
	return db.id
}

func Init() *DB {
	return &DB{
		id: ID,
	}
}
