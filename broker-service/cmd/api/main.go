package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go" //driver rabbitmq
)

const webPort = "80"

type Config struct {
	RabbitMQ *amqp.Connection
}

func main() {
	rabbitConn,err := connectRabbitMQ()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	
	defer rabbitConn.Close()
	app := Config{
		RabbitMQ: rabbitConn,
	}

	log.Printf("starting broker on port %s\n", webPort)

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectRabbitMQ() (*amqp.Connection,error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	//esperar rabbitmq estar ativo
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("rabbitmq not ready")
			counts++
		} else {
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil,err
		}

		//aumentar o tempo de espera para tentar connectar ao rabbitmq
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection,nil
}