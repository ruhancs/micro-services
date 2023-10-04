package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type RequestPayload struct {
	Action string `json:"action"`
	Auth AuthPayload `json:"auth"`
}

type AuthPayload struct {
	Email string `json:"email"`
	Password string `json:"password"`
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
	default:
		app.errorJson(w,errors.New("unknown action"))
	}
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