package amqp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Configuration struct {
	Name string
	URL  string
}

func NewConnection(config Configuration) *Connection {

	conn, err := amqp091.Dial(config.URL)

	if err != nil {
		panic(err)
	}

	channel, err := conn.Channel()

	if err != nil {
		panic(err)
	}

	return &Connection{
		name:    config.Name,
		channel: channel,
	}
}

type Connection struct {
	name    string
	channel *amqp091.Channel
}

func (q *Connection) SendMessage(queue string, key string, payload any) error {
	value, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	channel := q.channel

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

func (q *Connection) AddListener(queue string, handler func(ctx Context)) {
	channel := q.channel

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
				conn:     q,
				delivery: msg,
			}

			handler(ctx)
		}
	}()
}