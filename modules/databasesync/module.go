package databasesync

const ID = "databasesync"

type DB struct {
	id             string
	DatabaseConfig *Config
}

func (db *DB) SetupModule() {

}

func (db *DB) GetId() string {
	return db.id
}

func Init() *DB {
	return &DB{
		id: ID,
	}
}
