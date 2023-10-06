package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string `json:"action"`
	Auth AuthPayload `json:"auth"`
	Log LogPayload `json:"log,omitempty"`
	Mail MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From string `json:"from"`
	To string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error: false,
		Message: "Hit the broker",
	}

	_ = app.writeJson(w,http.StatusOK,payload)

}

func (app *Config) HandleSubmition(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.ReadJson(w,r, &requestPayload)
	fmt.Println(requestPayload)
	if err != nil {
		app.errorJson(w,err)
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logViaRPC(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:
		app.errorJson(w,errors.New("unknown action"))
	}
}

func(app *Config) LogEventViaRabbit(w http.ResponseWriter, log LogPayload) {
	err := app.PushToQueue(log.Name,log.Data)
	if err != nil {
		app.errorJson(w,err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via rabbitmq"

	app.writeJson(w,http.StatusAccepted,payload)
}

func(app *Config) PushToQueue(name,message string) error{
	emitter ,err := event.NewEventEmtter(app.RabbitMQ)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	jsonPayload,_ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(jsonPayload),"log.INFO")
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) logItem(w http.ResponseWriter, logDTO LogPayload) {
	jsonData,_ := json.MarshalIndent(logDTO, "", "\t")

	logServiceURL := "http://logger-service/log"
	request,err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w,err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response,err := client.Do(request)
	if err != nil {
		app.errorJson(w,err)
		return
	}
	defer response.Body.Close()
	
	if response.StatusCode != http.StatusAccepted {
		app.errorJson(w,err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "log sended"

	app.writeJson(w,http.StatusAccepted,payload)

}

func (app *Config) authenticate(w http.ResponseWriter, authDTO AuthPayload) {
	//criar json que sera enviado para o servico de authenticacao
	jsdonData,_ := json.MarshalIndent(authDTO, "", "\t")

	fmt.Println("CALL BROCKER AUTH")
	//chamar o servico de authenticacao, authentication-service nome do servico criado no docker compose
	request,err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsdonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}
	
	client := &http.Client{}
	response,err := client.Do(request)
	if err != nil {
		fmt.Println("error resp")
		fmt.Println("error resp")
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()
	
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid credential"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling auth service"))
		return
	}	
		
	var jsonFromAuthService jsonResponse
	//transformar a resposta do servico de authenticacao
	err = json.NewDecoder(response.Body).Decode(&jsonFromAuthService)
	if err != nil {
		fmt.Println("error new decoder")
		app.errorJson(w, err)
		return
	}
	
	if jsonFromAuthService.Error {
		fmt.Println("error fromAUTHSERVICE")
		app.errorJson(w,err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromAuthService.Data

	app.writeJson(w, http.StatusAccepted, payload)

}

func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData,_ := json.Marshal(msg)

	mailServiceURL := "http://mail-service/send"
	request,err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w,err)
		return
	}
	
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response,err := client.Do(request)
	if err != nil {
		app.errorJson(w,err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJson(w,errors.New("error to comunicate with mail service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "email sent to " +msg.To
	app.writeJson(w,http.StatusAccepted,payload)
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logViaRPC(w http.ResponseWriter, logPayload LogPayload) {
	//conexao com o servico de logger com a porta que esta rodando o rpc server
	clientRPC,err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJson(w,err)
		return
	}

	rpcPayload := RPCPayload{
		Name: logPayload.Name,
		Data: logPayload.Data,
	}

	var result string
	//funcao criada no servidor logger-service RPCServer
	err = clientRPC.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJson(w,err)
		return
	}

	var payload jsonResponse
	payload.Error=false
	payload.Message= result
	
	app.writeJson(w, http.StatusAccepted, payload)
}