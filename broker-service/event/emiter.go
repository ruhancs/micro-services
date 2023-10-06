package event

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go" //driver rabbitmq
)

type Emitter struct {
	connection *amqp.Connection
}

func (e *Emitter) Setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return declareExchange(channel)
}

//inserir evento no canal
func (e *Emitter) Push(event string, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	log.Println("pushing to channel")

	err = channel.PublishWithContext(
		context.Background(),
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte(event),
		},
	)
	if err != nil {
		return err
	}
	
	return nil
}

func NewEventEmtter(conn  *amqp.Connection) (Emitter,error) {
	emitter := Emitter{
		connection: conn,
	}

	err := emitter.Setup()
	if err != nil {
		return Emitter{},err
	}
	
	return emitter,nil
}