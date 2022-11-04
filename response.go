package amqp

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type Response struct {
	delivery amqp091.Delivery
}

func (res *Response) DecodeValue(v any) error {
	return json.Unmarshal(res.delivery.Body, v)
}
