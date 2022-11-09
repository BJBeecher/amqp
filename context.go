package amqp

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type Context struct {
	queue    string
	conn     *Connection
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
	payload := map[string]any{
		"code":    500,
		"message": message,
	}
	ctx.conn.Reply(ctx.delivery.ReplyTo, ctx.CorrelationId(), payload)
}

func (ctx *Context) AbortWithStatusMessage(code int, message string) {
	println("[AMQP-debug] Error:", message)
	payload := map[string]any{
		"code":    code,
		"message": message,
	}
	ctx.conn.Reply(ctx.delivery.ReplyTo, ctx.CorrelationId(), payload)
}

func (ctx *Context) Success(payload any) {
	ctx.conn.Reply(ctx.delivery.ReplyTo, ctx.CorrelationId(), payload)
}
