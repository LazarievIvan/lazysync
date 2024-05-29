package main

type ModuleHandler struct {
	modulesList []string
}

func getModuleHandler() ModuleHandler {
	return initModuleHandler()
}

func initModuleHandler() ModuleHandler {
	return ModuleHandler{
		modulesList: findModules(),
	}
}

func findModules() []string {
	return make([]string, 0)
}
