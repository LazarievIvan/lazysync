package server

import (
	"crypto"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	jsonrpc "github.com/gorilla/rpc/json"
	"io"
	manager "lazysync/application/service"
	"lazysync/modules"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const Type = "server"

const port string = ":8080"

type Server struct {
	Configuration  *manager.AppConfiguration
	ActiveSessions map[string][]byte
}

func (s *Server) GetType() string {
	return Type
}

func (s *Server) SetMode(module modules.Module) {
	s.Configuration.Module = module.GetId()
	s.Configuration.ModuleSpecificConfig = module.GetConfigurationValues()
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
	manager.SaveConfiguration(s.Configuration)
	// Generate module specific Configuration.
	s.GenerateKeys(2)
}

func (s *Server) Run() {
	fmt.Println("Starting server...")
	s.StartServer()
}

func (s *Server) AuthorizeUserWithKey(username string, signature []byte) bool {
	userPubKey := manager.ReadPublicKey(username)
	hashedUsername := sha256.Sum256([]byte(username))
	err := rsa.VerifyPKCS1v15(userPubKey, crypto.SHA256, hashedUsername[:], signature)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (s *Server) AuthorizeUserWithToken(username string, token string) bool {
	err := s.verifyToken(username, token)
	if err != nil {
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

func (s *Server) createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
	rand.New(rand.NewSource(time.Now().UnixNano()))
	numberOfBytes := rand.Intn(256-128+1) + 128
	secret := manager.GenerateRandomBytesSequence(numberOfBytes)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	s.ActiveSessions[username] = secret
	return tokenString, nil
}

func (s *Server) verifyToken(username string, tokenString string) error {
	secret := s.ActiveSessions[username]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (s *Server) Authorize(r *http.Request, args *manager.AuthenticationArgs, reply *manager.AuthenticationResponse) error {
	var response manager.AuthenticationResponse
	if args.Token == nil {
		return errors.New("no token provided")
	}
	token := args.Token
	authorized, err := s.performAuthentication(token)
	if !authorized || err != nil {
		return errors.New("not authorized")
	}
	if token.TokenType == manager.TokenTypeKey {
		jwtToken, err := s.createToken(token.Username)
		if err != nil {
			return err
		}
		response.Object = jwtToken
	}
	response.Status = http.StatusOK
	*reply = response
	return nil
}

func (s *Server) performAuthentication(token *manager.AuthenticationToken) (bool, error) {
	authorized := false
	switch token.TokenType {
	case manager.TokenTypeKey:
		authorized = s.AuthorizeUserWithKey(token.Username, token.Token)
	case manager.TokenTypeJWT:
		authorized = s.AuthorizeUserWithToken(token.Username, string(token.Token))
		if !authorized {
			return false, errors.New("invalid token")
		}
	}
	return authorized, nil
}

func (s *Server) Synchronize(r *http.Request, args *manager.SynchronizationArgs, reply *manager.SynchronizationResponse) error {
	var response manager.SynchronizationResponse
	authorized, err := s.performAuthentication(args.Token)
	if !authorized || err != nil {
		return errors.New("not authorized, please re-run application")
	}
	if s.Configuration.Module != args.Module {
		return errors.New("enabled module does not match given module: " + args.Module)
	}
	moduleName := args.Module
	moduleInstance := modules.InitModuleHandler()
	module, err := moduleInstance.GetModuleByName(moduleName)
	if err != nil {
		return errors.New("module not found: " + moduleName)
	}
	module.SetConfiguration(s.Configuration.ModuleSpecificConfig)
	syncResponse := module.Sync()
	response.Status = http.StatusOK
	response.Object = &syncResponse
	*reply = response
	return nil
}

func (s *Server) StartServer() {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(jsonrpc.NewCodec(), "application/json")
	err := rpcServer.RegisterService(s, "")
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()
	router.Handle("/", rpcServer)
	// Registered module-specific routers, if any.
	moduleName := s.Configuration.Module
	moduleInstance := modules.InitModuleHandler()
	module, err := moduleInstance.GetModuleByName(moduleName)
	if err != nil {
		log.Fatal("module not found: " + moduleName)
	}
	module.SetConfiguration(s.Configuration.ModuleSpecificConfig)
	if module, ok := module.(modules.WebServiceModule); ok {
		err = rpcServer.RegisterService(module, "")
		module.RegisterAsWebService(router, rpcServer)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Started on port", port)
	fmt.Println("To close connection CTRL+C")
	err = http.ListenAndServe(port, router)
	if err != nil {
		log.Fatal(err)
	}
}
