package amqp

import (
	"encoding/json"
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
	return json.Unmarshal(ctx.delivery.Body, v)
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
