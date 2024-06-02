package service

const TokenTypeKey = "key"

const TokenTypeJWT = "jwt"

type Response struct {
	Result *BaseResponse `json:"result"`
	Error  interface{}   `json:"error"`
	ID     string        `json:"id"`
}

type BaseResponse struct {
	Status int         `json:"status"`
	Object interface{} `json:"object"`
}

type AuthenticationToken struct {
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	Token     []byte `json:"access_token"`
}

type AuthenticationArgs struct {
	Token *AuthenticationToken `json:"token"`
}

type AuthenticationRequest struct {
	Method string               `json:"method"`
	Params []AuthenticationArgs `json:"params"`
	Id     string               `json:"id"`
}

type AuthenticationResponse struct {
	Status int    `json:"status"`
	Object string `json:"token"`
}

func NewAuthenticationResponse(response *BaseResponse) *AuthenticationResponse {
	authenticationResponse := new(AuthenticationResponse)
	authenticationResponse.Status = response.Status
	if _, ok := response.Object.(string); ok {
		authenticationResponse.Object = response.Object.(string)
	}
	return authenticationResponse
}

type SynchronizationArgs struct {
	Module string               `json:"module"`
	Token  *AuthenticationToken `json:"token"`
}

type SynchronizationRequest struct {
	Method string                `json:"method"`
	Params []SynchronizationArgs `json:"params"`
	Id     string                `json:"id"`
}

type SynchronizationResponse struct {
	Status int         `json:"status"`
	Object *SyncObject `json:"object"`
}

func NewAuthenticationRequest() *AuthenticationRequest {
	request := new(AuthenticationRequest)
	request.Method = "Server.Authorize"
	return request
}

func NewSynchronizationRequest() *SynchronizationRequest {
	request := new(SynchronizationRequest)
	request.Method = "Server.Synchronize"
	return request
}
