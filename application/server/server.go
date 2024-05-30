package server

import (
	"crypto"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	manager "lazysync/application/service"
	"lazysync/application/web"
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
	Configuration *manager.AppConfiguration
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

func (s *Server) AuthorizeUser(username string, signature []byte) bool {
	userPubKey := manager.ReadPublicKey(username)
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

func (s *Server) Authorize(accessHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := web.GetUserFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		authorized := s.AuthorizeUser(user.Username, user.Signature)
		if !authorized {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Authorize via server.
		accessHandler.ServeHTTP(w, r)
	})
}

func (s *Server) StartServer() {
	authHandler := http.HandlerFunc(GrantAccess)
	mux := http.NewServeMux()
	mux.Handle("/", s.Authorize(authHandler))
	//mux.Handle("/sync", http.HandlerFunc(ProcessSync))
	log.Println("Started on port", port)
	fmt.Println("To close connection CTRL+C")
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func GrantAccess(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Authorized"))
	if err != nil {
		log.Fatal(err)
	}
}
