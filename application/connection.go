package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"lazysync/modules"
	"log"
	"net/http"
)

// @todo Implement JSON-RPC.
const port string = ":8080"

type User struct {
	Username  string
	Signature []byte
}

func Authorize(server *Server, accessHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		authorized := server.AuthorizeUser(user.Username, user.Signature)
		if !authorized {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Authorize via server.
		accessHandler.ServeHTTP(w, r)
	})
}

func getUserFromRequest(r *http.Request) (*User, error) {
	var user User

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GrantAccess(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Authorized"))
	if err != nil {
		log.Fatal(err)
	}
}

func ProcessSync(w http.ResponseWriter, r *http.Request) {
	// @todo change the stub.
	fmt.Println("Sync Process")
}

func StartServer(server *Server) {
	authHandler := http.HandlerFunc(GrantAccess)
	mux := http.NewServeMux()
	mux.Handle("/", Authorize(server, authHandler))
	mux.Handle("/sync", http.HandlerFunc(ProcessSync))
	log.Println("Started on port", port)
	fmt.Println("To close connection CTRL+C")
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func Connect(username string, signature []byte) (int, error) {
	user := User{Username: username, Signature: signature}
	jsonData, err := json.Marshal(user)

	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewBuffer(jsonData))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusAccepted, err
	}
	return resp.StatusCode, nil
}

func RunSync(module modules.Module) {
	// @todo prepare sync and run request.
	_, err := http.NewRequest("GET", "http://localhost:8080/sync", nil)
	if err != nil {
		panic(err)
	}
}
