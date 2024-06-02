package client

import (
	"crypto"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	manager "lazysync/application/service"
	"lazysync/application/web"
	"lazysync/modules"
	"os"
)

const Type = "client"

type Client struct {
	Configuration *manager.AppConfiguration
	JWTToken      string
}

func (c *Client) GetType() string {
	return Type
}

func (c *Client) SetMode(module modules.Module) {
	c.Configuration.Mode = c.GetType()
	c.Configuration.Module = module.GetId()
}

func (c *Client) Setup() {
	fmt.Println("Setting up client...")
	username, err := c.ScanUsername()
	if err != nil {
		panic(err)
	}
	c.Configuration.Username = username
	manager.SaveConfiguration(c.Configuration)
}

func (c *Client) Run() {
	fmt.Println("Starting client...")
	username := c.Configuration.Username
	key := manager.ReadPrivateKey(username)
	hashedUsername := sha256.Sum256([]byte(username))
	signature, _ := rsa.SignPKCS1v15(cryptoRand.Reader, key, crypto.SHA256, hashedUsername[:])
	response, err := web.Login(username, signature)
	if err != nil {
		panic(err)
	}
	c.JWTToken = response.Object
	moduleName := c.Configuration.Module
	moduleInstance := modules.InitModuleHandler()
	module, err := moduleInstance.GetModuleByName(moduleName)
	if err != nil {
		panic(err)
	}
	expectedObject := module.GetSyncObjectInstance()
	syncResponse, err := web.Sync(username, c.JWTToken, c.Configuration.Module, expectedObject)
	if err != nil {
		panic(err)
	}
	module.ExecuteCommands(*syncResponse)
}

func (c *Client) ScanUsername() (string, error) {
	keys, err := os.ReadDir(manager.KeyBasePath)
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "", errors.New("no username found, please contact administrator")
	}
	if len(keys) > 1 {
		return "", errors.New("multiple username found, please contact administrator")
	}
	return keys[0].Name(), nil
}
