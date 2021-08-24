package core

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const EventsExchange = "events"

type RabbitMQEventNetwork struct {
	rabbitMqChannel       *amqp.Channel
	eventReceivedCallBack EventHandler
	logger                *log.Entry
}

type ConnexionDetails struct {
	Username string
	Password string
	Host     string
	Port     string
}

func NewRabbitMQEventNetwork(connDetails ConnexionDetails, mainLogger *log.Entry) *RabbitMQEventNetwork {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		connDetails.Username,
		connDetails.Password,
		connDetails.Host,
		connDetails.Port,
	))
	if err != nil {
		mainLogger.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		mainLogger.Fatalf("failed to open a Channel: %v", err)
	}

	// setting up the different exchange
	err = ch.ExchangeDeclare(
		EventsExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil)

	if err != nil {
		mainLogger.Fatalf("could not declare the events exchange: %v", err)
	}

	return &RabbitMQEventNetwork{
		rabbitMqChannel: ch,
		logger:          mainLogger.WithField("component", "event-network"),
	}
}

func (r *RabbitMQEventNetwork) BroadcastEvent(event *Event) {
	if event.Receiver == "" {
		event.Receiver = "*"
	}

	data, err := json.Marshal(event)
	if err != nil {
		r.logger.Errorf("could not marshal event: %v", err)
	}

	err = r.rabbitMqChannel.Publish(
		EventsExchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})
	if err != nil {
		r.logger.Errorf("could not send event: %v", err)
	}
}

func (r *RabbitMQEventNetwork) SendEventTo(receiver string, event *Event) {
	event.Receiver = receiver
	r.BroadcastEvent(event)
}

func (r *RabbitMQEventNetwork) SetReceivedEventCallback(handler EventHandler) {
	r.eventReceivedCallBack = handler
}

func (r *RabbitMQEventNetwork) StartListeningForEvents() {
	q, err := r.rabbitMqChannel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		r.logger.Fatalf("failed to declare a queue: %v", err)
	}
	r.logger.Debugf("queue %s declared", q.Name)

	err = r.rabbitMqChannel.QueueBind(
		q.Name,         // queue name
		"",             // routing key
		EventsExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		r.logger.Fatalf("failed to bind a queue: %v", err)
	}
	r.logger.Debugf("Successfully bound to queue %s", q.Name)

	msgs, err := r.rabbitMqChannel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		r.logger.Fatalf("failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			var event Event
			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				r.logger.Warnf("could not unmarshal event: %v", err)
			}
			r.eventReceivedCallBack(&event)
		}
	}()

	r.logger.Info("Listening for events...")
}
