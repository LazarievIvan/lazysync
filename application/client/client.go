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
	"net/http"
	"os"
)

const Type = "client"

type Client struct {
	Configuration *manager.AppConfiguration
}

func (c *Client) GetType() string {
	return Type
}

func (c *Client) SetMode(module modules.Module) {
	c.Configuration.Module = module.GetId()
}

func (c *Client) Setup() {
	fmt.Println("Setting up client...")
	c.Configuration.Mode = c.GetType()
	username, err := c.ScanUsername()
	if err != nil {
		panic(err)
	}
	c.Configuration.Username = username
	manager.SaveConfiguration(c.Configuration)
}

func (c *Client) Run() {
	fmt.Println("Starting server...")
	username := c.Configuration.Username
	key := manager.ReadPrivateKey(username)
	hashedUsername := sha256.Sum256([]byte(username))
	signature, _ := rsa.SignPKCS1v15(cryptoRand.Reader, key, crypto.SHA256, hashedUsername[:])
	status, err := web.Connect(username, signature)
	if err != nil {
		panic(err)
	}
	if status == http.StatusOK {
		moduleHandler := modules.InitModuleHandler()
		_, err := moduleHandler.GetModuleByName(c.Configuration.Mode)
		if err != nil {
			panic(err)
		}
	}
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
