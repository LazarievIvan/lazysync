package databasesync

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c *Config) GetConfigMap() map[string]string {
	return map[string]string{
		"host":     c.Host,
		"port":     c.Port,
		"username": c.Username,
		"password": c.Password,
		"database": c.Database,
	}
}

func saveToFile(cfg *Config) error {
	return nil
}

func createConfig() {

}
