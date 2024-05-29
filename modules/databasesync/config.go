package databasesync

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func saveToFile(cfg *Config) error {
	return nil
}

func createConfig() {

}
