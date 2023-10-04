package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort = "80"
	rpcPort = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRPCPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient,err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// contexto para desconectar, timeout de 15 segundos
	ctx,cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	//fechar conexao
	defer func () {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	go app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}


func connectToMongo() (*mongo.Client,error) {
	clientOpt := options.Client().ApplyURI(mongoURL)
	clientOpt.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	conn,err := mongo.Connect(context.TODO(),clientOpt)
	if err != nil {
		log.Println("Error to connect to mongoDB: ",err)
		return nil,err
	}

	return conn,nil
}