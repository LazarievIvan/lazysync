package service

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
)

const KeyBasePath = "private/keys/"

func ReadPublicKey(username string) *rsa.PublicKey {
	path := KeyBasePath + username + "/key.rsa.pub"
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

func ReadPrivateKey(username string) *rsa.PrivateKey {
	path := KeyBasePath + username + "/key.rsa"
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
