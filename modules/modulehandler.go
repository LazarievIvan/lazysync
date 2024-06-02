package modules

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"lazysync/application/service"
	"lazysync/modules/filesystem"
)

type Module interface {
	GetId() string
	SetupModule()
	GetConfigurationValues() interface{}
	SetConfiguration(configuration interface{})
	Sync() service.SyncObject
	GetSyncObjectInstance() service.SyncObject
	ExecuteCommands(object service.SyncObject)
}

type WebServiceModule interface {
	RegisterAsWebService(router *mux.Router, server *rpc.Server)
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
	return map[string]Module{
		filesync.GetId(): filesync,
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
