package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/tidwall/gjson"
	"io"
	"lazysync/application/service"
	"net/http"
)

const DefaultUrl = "http://localhost:8080"

const MethodGet = "GET"

type User struct {
	Username  string
	Signature []byte
}

type Request struct {
	Method func() `json:"method,omitempty"`
	ID     string `json:"id,omitempty"`
}

func Login(username string, signature []byte) (*service.AuthenticationResponse, error) {
	authentication := service.AuthenticationToken{Username: username, TokenType: service.TokenTypeKey, Token: signature}
	connectionArguments := service.AuthenticationArgs{Token: &authentication}
	authenticationRequest := service.NewAuthenticationRequest()
	authenticationRequest.Params = append(authenticationRequest.Params, connectionArguments)
	authenticationRequest.Id = "1"
	jsonData, err := json.Marshal(authenticationRequest)
	resp, err := SendJsonRequest(http.MethodPost, DefaultUrl, jsonData)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	response := &service.BaseResponse{
		Status: 0,
		Object: nil,
	}
	parseResult := gjson.Get(string(respBody), "result")
	response.Status = int(parseResult.Get("status").Int())
	response.Object = parseResult.Get("token").String()
	authResponse := service.NewAuthenticationResponse(response)
	if err != nil {
		return authResponse, err
	}
	return authResponse, nil
}

func Sync(username string, token string, module string, result service.SyncObject) (*service.SyncObject, error) {
	arguments := service.SynchronizationArgs{
		Module: module,
		Token:  &service.AuthenticationToken{Username: username, TokenType: service.TokenTypeJWT, Token: []byte(token)},
	}
	request := service.NewSynchronizationRequest()
	request.Params = append(request.Params, arguments)
	request.Id = "2"
	jsonData, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	resp, err := SendJsonRequest(http.MethodPost, DefaultUrl, jsonData)
	if err != nil {
		panic(err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	parseResult := gjson.Get(string(respBody), "result")
	status := int(parseResult.Get("status").Int())
	objectResponse := parseResult.Get("object").String()
	result.ParseResponse(objectResponse)
	if status != http.StatusOK {
		return &result, errors.New("request accepted")
	}
	return &result, nil
}

func SendJsonRequest(method string, url string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	return resp, err
}
