package event

import (
	amqp "github.com/rabbitmq/amqp091-go" //driver rabbitmq
)

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", //nome da exchange
		"topic",      //tipo
		true,         //duravel
		false,        //auto deletado eventos
		false,        //sera usado somente internamente
		false,        // sem espera
		nil,
	)
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue,error) {
	return ch.QueueDeclare(
		"", //nome
		false, //duravel
		false, // delete quando nao estiver usando
		true, // se o canal e exclusivo
		false, // sem espera
		nil,
	)
}