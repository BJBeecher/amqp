package amqp

type Payload struct {
	Failure *Failure `json:"failure"`
	Data    *any     `json:"data"`
}
