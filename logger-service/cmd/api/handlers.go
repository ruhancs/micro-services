package main

import (
	"logger-service/data"
	"net/http"
)

type ReqPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload ReqPayload

	_ = app.ReadJson(w,r, &requestPayload)

	//inserir os dados da request no database, na colecao logs
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}
	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.errorJson(w,err)
		return
	}

	resp := jsonResponse{
		Error: false,
		Message: "salved log",
	}

	app.writeJson(w, http.StatusAccepted,resp)
}