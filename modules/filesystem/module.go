package filesystem

import (
	"lazysync/modules/filesystem/cmd"
)

const ID = "filesystem"

type FileSync struct {
	id            string
	Configuration FileSyncConfig
}

type FileSyncConfig struct {
	Files []string // Files list.
}

func (f *FileSync) GetId() string {
	return f.id
}

func (f *FileSync) SetupModule() {
	f.Configuration.Files = cmd.Setup()
}

func (f *FileSync) GetConfigurationValues() any {
	return f.Configuration
}

func (f *FileSync) Sync() {}

func Init() *FileSync {
	return &FileSync{id: ID, Configuration: FileSyncConfig{}}
}
