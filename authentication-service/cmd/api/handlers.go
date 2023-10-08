package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.ReadJson(w,r, &requestPayload)
	if err != nil {
		app.errorJson(w,err, http.StatusBadRequest)
		return
	}

	user,err := app.Repo.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJson(w,errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}
	
	valid,err := app.Repo.PasswordMatches(requestPayload.Password, *user)
	if err != nil || !valid {
		app.errorJson(w,errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}
	
	//enviar o evento para o servico de logs
	err = app.LogRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil || !valid {
		app.errorJson(w,err, http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: fmt.Sprintf("Logged in user"),
		Data: user,
	}

	app.writeJson(w, http.StatusAccepted, payload)
	
}

func (app *Config) LogRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	entry.Name = name
	entry.Data = data

	jsonData,_ := json.MarshalIndent(entry,"", "\t")
	
	logServiceURL := "http://logger-service/log"
	request,err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	_,err = app.Client.Do(request)
	if err != nil {
		return err
	}
	
	return nil
}