package amqp

import (
	"encoding/json"
	"errors"

	"github.com/rabbitmq/amqp091-go"
)

type Response struct {
	delivery amqp091.Delivery
}

func (res *Response) DecodeValue(v any) error {
	var payload Payload

	err := json.Unmarshal(res.delivery.Body, &payload)

	if err != nil {
		return err
	}

	if payload.Data == nil {
		return errors.New("[AMQP-debug] Error: expected data in payload")
	}

	data, err := json.Marshal(&payload.Data)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}
