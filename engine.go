package amqp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Configuration struct {
	Name string
	URL  string
}

func New(config Configuration) *Engine {

	conn, err := amqp091.Dial(config.URL)

	if err != nil {
		panic(err)
	}

	channel, err := conn.Channel()

	if err != nil {
		panic(err)
	}

	return &Engine{
		name:       config.Name,
		connection: conn,
		channel:    channel,
	}
}

type Engine struct {
	name       string
	connection *amqp091.Connection
	channel    *amqp091.Channel
}

func (e *Engine) Publish(queue string, key string, data any) error {
	payload := Payload{Data: &data}
	value, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	channel := e.channel

	que, err := channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		ctx,      // context
		"",       // exchange
		que.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        value,
		},
	)

	if err != nil {
		println("[AMQP-debug] Error -", err)
	} else {
		println("[AMQP-debug] Message sent to queue:", queue)
	}

	return err
}

func (e *Engine) SendRequest(routingKey string, data any) (*Response, error) {
	payload := Payload{Data: &data}
	value, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	channel := e.channel

	queue, err := channel.QueueDeclare(
		"",    // name - default
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return nil, err
	}

	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)

	if err != nil {
		return nil, err
	}

	correlationId, err := uuid.NewRandom()

	if err != nil {
		println("[AMQP-debug] Error -", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		ctx,        // context
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp091.Publishing{
			ContentType:   "application/json",
			Body:          value,
			CorrelationId: correlationId.String(),
			ReplyTo:       queue.Name,
		},
	)

	if err != nil {
		println("[AMQP-debug] Error -", err)
		return nil, err
	}

	var response Response

	for delivery := range msgs {
		if delivery.CorrelationId == correlationId.String() {
			response = Response{delivery: delivery}
			break
		}
	}

	var body Payload
	err = response.DecodeValue(&body)

	if err != nil {
		println("[AMQP-debug] Error -", err)
		return nil, err
	}

	if body.Failure != nil {
		errMessage := fmt.Sprintf("%v: %s", body.Failure.Code, body.Failure.Message)
		return nil, errors.New(errMessage)
	}

	return &response, nil
}

func (e *Engine) Reply(queue string, correlationId string, data any) {
	payload := Payload{Data: &data}
	value, err := json.Marshal(payload)

	if err != nil {
		println("[AMQP-debug] Error -", err)
		return
	}

	channel := e.channel

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		ctx,   // context
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType:   "application/json",
			Body:          value,
			CorrelationId: correlationId,
		},
	)

	if err != nil {
		println("[AMQP-debug] Error -", err)
	} else {
		println("[AMQP-debug] Message sent to queue:", queue)
	}
}

func (e *Engine) AddListener(queue string, handler func(ctx Context)) {
	channel := e.channel

	que, err := channel.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		panic(err)
	}

	msgs, err := channel.Consume(
		que.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)

	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range msgs {
			println("[AMQP-debug] Message recieved at queue:", queue)

			ctx := Context{
				queue:    queue,
				engine:   e,
				delivery: msg,
			}

			handler(ctx)
		}
	}()
}
