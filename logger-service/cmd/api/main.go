package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort = "80"
	rpcPort = "5001"
	mongoURL = "mongodb://mongo:27017"
	//mongoURL = "mongodb://localhost:27017"
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

	//rpc server, registrar rpc server
	err = rpc.Register(new(RPCServer))
	go app.rpcListen()

	//iniciar o servidor Grpc
	go app.GrpcListen()

	log.Println("starting log service on port: ",webPort)
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func (app *Config) rpcListen() error {
	log.Println("starting rpc server on port: ",webPort)
	listen,err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s",rpcPort))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn,err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
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
	log.Println("connected to mongo")

	return conn,nil
}