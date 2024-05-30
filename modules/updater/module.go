package updater

const ID = "updater"

type Update struct {
	id            string
	Configuration *UpdateConfig
}

type UpdateConfig struct {
	AppName    string `toml:"name"`
	AppVersion int    `yaml:"version"`
}

func (up *Update) GetId() string {
	return up.id
}

func (up *Update) SetupModule() {
}

func (up *Update) GetConfigurationValues() any {
	return up.Configuration
}

func (up *Update) Sync() {}

func Init() *Update {
	return &Update{id: ID}
}
