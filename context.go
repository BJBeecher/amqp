package amqp

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rabbitmq/amqp091-go"
)

type Context struct {
	queue    string
	engine   *Engine
	delivery amqp091.Delivery
}

func (ctx *Context) CorrelationId() string {
	return ctx.delivery.CorrelationId
}

func (ctx *Context) DecodeValue(v any) error {
	var payload Payload

	err := json.Unmarshal(ctx.delivery.Body, &payload)

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

func (ctx *Context) AbortWithError(err error) {
	message := err.Error()
	println("[AMQP-debug] Error:", message)
	payload := Failure{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
	ctx.engine.Reply(ctx.delivery.ReplyTo, ctx.CorrelationId(), payload)
}

func (ctx *Context) AbortWithStatusMessage(code int, message string) {
	println("[AMQP-debug] Error:", message)
	payload := Failure{
		Code:    code,
		Message: message,
	}
	ctx.engine.Reply(ctx.delivery.ReplyTo, ctx.CorrelationId(), payload)
}

func (ctx *Context) Success(payload any) {
	ctx.engine.Reply(ctx.delivery.ReplyTo, ctx.CorrelationId(), payload)
}
