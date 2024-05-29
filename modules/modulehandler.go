package modules

import (
	"fmt"
	"lazysync/modules/databasesync"
)

type Module interface {
	GetId() string
	SetupModule()
	Sync()
}

type ModuleConfig interface {
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
	db := databasesync.Init()
	return map[string]Module{
		db.GetId(): db,
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
	err := fmt.Errorf("Module %s not found", name)
	return nil, err
}