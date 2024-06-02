package application

import (
	"lazysync/application/client"
	"lazysync/application/server"
	"lazysync/application/service"
	"lazysync/modules"
)

type App interface {
	Setup()
	Run()
	GetType() string
	SetMode(module modules.Module)
}

func InitFromConfig() App {
	config := service.LoadConfiguration()
	switch config.Mode {
	case server.Type:
		return &server.Server{Configuration: config, ActiveSessions: map[string][]byte{}}
	case client.Type:
		return &client.Client{Configuration: config, JWTToken: ""}
	}
	return nil
}
