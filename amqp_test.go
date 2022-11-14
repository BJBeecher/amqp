package amqp

import "testing"

func TestRequest(t *testing.T) {
	type TestPayload struct {
		Text string `json:"message"`
	}
	engine := New(Configuration{URL: "amqp://guest:guest@localhost:5672"})

	engine.AddListener("test", func(ctx Context) {
		var req TestPayload
		err := ctx.DecodeValue(&req)

		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		ctx.Success(req)
	})

	res, err := engine.SendRequest("test", TestPayload{Text: "hello world"})

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	var payload TestPayload
	err = res.DecodeValue(&payload)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
