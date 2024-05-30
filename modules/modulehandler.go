package modules

import (
	"fmt"
	"lazysync/modules/filesystem"
)

type Module interface {
	GetId() string
	SetupModule()
	GetConfigurationValues() any
	Sync()
}

type ModuleHandler struct {
	ModulesList map[string]Module
}

func InitModuleHandler() *ModuleHandler {
	return &ModuleHandler{
		ModulesList: findModules(),
	}
}

func findModules() map[string]Module {
	filesync := filesystem.Init()
	//upgrader := updater.Init()
	return map[string]Module{
		filesync.GetId(): filesync,
		//upgrader.GetId(): upgrader,
	}
}

func (mh *ModuleHandler) GetModuleNamesList() []string {
	var moduleNames []string
	for name := range mh.ModulesList {
		moduleNames = append(moduleNames, name)
	}
	return moduleNames
}

func (mh *ModuleHandler) GetModuleByName(name string) (Module, error) {
	if module, ok := mh.ModulesList[name]; ok {
		return module, nil
	}
	err := fmt.Errorf("module %s not found", name)
	return nil, err
}
