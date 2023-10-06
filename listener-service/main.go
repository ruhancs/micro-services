package main

import (
	"fmt"
	"listener-service/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go" //driver rabbitmq
)

func main() {
	rabbitConn,err := connectRabbitMQ()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	//pegar as menssagens
	log.Println("listening and consuming rabbitmq messages")

	//criar consumer
	consumer,err := event.NewConsumer(rabbitConn)
	if err != nil {
		panic(err)
	}

	//observar a queue e consumir os eventos
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.WARNING"})
	if err != nil {
		log.Println(err)
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
			fmt.Println(err)
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