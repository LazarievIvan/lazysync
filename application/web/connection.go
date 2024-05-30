package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// @todo Implement JSON-RPC.

type User struct {
	Username  string
	Signature []byte
}

func GetUserFromRequest(r *http.Request) (*User, error) {
	var user User

	// Try to decode the request body into the struct. If there is an error,
	// respond to the server with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func ProcessSync(w http.ResponseWriter, r *http.Request) {
	// @todo change the stub.
	fmt.Println("Sync Process")
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

func RunSync() {
	// @todo prepare sync and run request.
	_, err := http.NewRequest("GET", "http://localhost:8080/sync", nil)
	if err != nil {
		panic(err)
	}
}
