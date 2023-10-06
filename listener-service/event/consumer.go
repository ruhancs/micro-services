package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go" //driver rabbitmq
)

type Consumer struct {
	conn *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer,error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{},err
	}

	return consumer,nil
}

func (c *Consumer) setup() error {
	channel,err := c.conn.Channel()
	if err != nil {
		return err
	}

	//cria a exchange
	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

//pegar dados de um topico
func (c *Consumer) Listen(topics []string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	//criar queue randomica
	q,err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _,value := range topics {
		//adicionar o canal com a queue
		ch.QueueBind(
			q.Name,
			value,
			"logs_topic",
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	//consumeir as msg
	messages,err := ch.Consume(q.Name, "", true, false,false,false,nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages{
			var payload Payload
			//inserir o conteudo ds menssagens no payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Println("Wating for messages for message [exchange, queue] [logs_topic]")
	//rodar para sempre
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auht":
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func logEvent(payload Payload) error {
	jsonData,_ := json.MarshalIndent(payload, "", "\t")

	logServiceURL := "http://logger-service/log"
	request,err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response,err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	
	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}