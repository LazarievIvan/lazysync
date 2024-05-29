package application

import (
	"crypto"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"lazysync/modules"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const ConfigFile = "config.yaml"

const ServerType = "server"

const ClientType = "client"

type App interface {
	Setup()
	Run()
	GetType() string
	SetMode(module modules.Module)
}

type AppConfiguration struct {
	Mode                 string            `yaml:"mode"`
	Username             string            `yaml:"username"`
	Module               string            `yaml:"module"`
	ModuleSpecificConfig map[string]string `yaml:"config"`
}

type Server struct {
	Configuration *AppConfiguration
}

type Client struct {
	Configuration *AppConfiguration
}

func InitFromConfig() App {
	var config AppConfiguration
	yamlFile, err := os.ReadFile(ConfigFile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	fmt.Println(config.Mode)
	switch config.Mode {
	case ServerType:
		return &Server{Configuration: &config}
	case ClientType:
		return &Client{Configuration: &config}
	}
	return nil
}

func (s *Server) GetType() string {
	return ServerType
}

func (s *Server) SetMode(module modules.Module) {
	s.Configuration.Module = module.GetId()
	s.Configuration.ModuleSpecificConfig = module.GetConfiguration()
}

func (s *Server) Setup() {
	fmt.Println("Setting up server...")
	s.Configuration.Mode = s.GetType()
	s.Configuration.Username = s.GetType()
	/*
		App Configuration is saved to .yaml file.
		Step 1: generate base Configuration with: app mode
		Step 2: generate module specific Configuration starting with: module
	*/
	saveConfiguration(s.Configuration)
	// Generate module specific Configuration.
	s.GenerateKeys(2)
}

func (s *Server) Run() {
	fmt.Println("Starting server...")
	StartServer(s)
}

func (s *Server) AuthorizeUser(username string, signature []byte) bool {
	userPubKey := readPublicKey(username)
	hashedUsername := sha256.Sum256([]byte(username))
	err := rsa.VerifyPKCS1v15(userPubKey, crypto.SHA256, hashedUsername[:], signature)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (s *Server) GenerateKeys(usersAmount int) {
	fmt.Println("Generating crypto keys...")
	bitSize := 4096
	for i := 0; i <= usersAmount; i++ {
		ownKeyGeneration := i == usersAmount
		key, err := rsa.GenerateKey(cryptoRand.Reader, bitSize)
		if err != nil {
			panic(err)
		}
		var path string
		if !ownKeyGeneration {
			username := s.generateUsername()
			path = "private/keys/" + username
		} else {
			path = "private/keys/server"
		}
		err = os.MkdirAll(path, 0750)
		if err != nil {
			panic(err)
		}
		s.saveToFile(key, path)
	}
}

func (s *Server) saveToFile(key *rsa.PrivateKey, path string) {
	filename := "/key"
	// Encode private key to PKCS#1 ASN.1 PEM.
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	// Encode public key to PKCS#1 ASN.1 PEM.
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
		},
	)
	// Write private key to file.
	if err := os.WriteFile(path+filename+".rsa", keyPEM, 0700); err != nil {
		panic(err)
	}

	// Write public key to file.
	if err := os.WriteFile(path+filename+".rsa.pub", pubPEM, 0755); err != nil {
		panic(err)
	}
	return
}

func (_ *Server) generateUsername() string {
	// Read local dictionary.
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	// Get words.
	words := strings.Split(string(bytes), "\n")
	// Get random words.
	rand.New(rand.NewSource(time.Now().UnixNano()))
	totalWords := 2
	var sb strings.Builder
	for i := 0; i < totalWords; i++ {
		word := words[rand.Intn(len(words)+1)]
		if strings.Contains(word, "'") {
			word = strings.SplitN(word, "'", 2)[0]
		}
		sb.WriteString(word)
		if i != totalWords-1 {
			sb.WriteString("_")
		}
	}
	return sb.String()
}

func (c *Client) GetType() string {
	return ClientType
}

func (c *Client) SetMode(module modules.Module) {
	c.Configuration.Mode = c.GetType()
	c.Configuration.Module = module.GetId()
}

func (c *Client) Setup() {
	fmt.Println("Setting up client...")
	c.Configuration.Username = "synergy_animators"
	saveConfiguration(c.Configuration)
}

func (c *Client) Run() {
	fmt.Println("Starting client...")
	username := c.Configuration.Username
	key := readPrivateKey(username)
	hashedUsername := sha256.Sum256([]byte(username))
	signature, _ := rsa.SignPKCS1v15(cryptoRand.Reader, key, crypto.SHA256, hashedUsername[:])
	status, err := Connect(username, signature)
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

func readPublicKey(username string) *rsa.PublicKey {
	path := "private/keys/" + username + "/key.rsa.pub"
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(bytes)
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key
}

func readPrivateKey(username string) *rsa.PrivateKey {
	path := "private/keys/" + username + "/key.rsa"
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(bytes)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key
}

func saveConfiguration(configuration *AppConfiguration) {
	yamlContents, err := yaml.Marshal(configuration)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(ConfigFile, yamlContents, 0644)
	if err != nil {
		panic(err)
	}
}
