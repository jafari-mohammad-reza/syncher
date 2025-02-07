package share

import (
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsConn struct {
	conn *nats.Conn
}

func NewNatsConn(url string) *NatsConn {
	conn, err := nats.Connect(url)
	if err != nil {
		slog.Error("Nats connection", "err", err.Error())
		os.Exit(1)
	}
	return &NatsConn{
		conn,
	}
}

func (nc *NatsConn) SubscribeToSubject(sbj string) (*nats.Subscription, error) {
	sub, err := nc.conn.QueueSubscribeSync(sbj, "servers")
	if err != nil {
		slog.Error("NatsConn SubscribeSync", "err", err.Error())
	}
	return sub, nil
}

func (nc *NatsConn) PublishToSubject(sbj string, data []byte) error {
	err := nc.conn.Publish(sbj, data)
	if err != nil {
		slog.Error("NatsConn PublishToSubject", "err", err.Error())
		return err
	}
	return nil
}

func (nc *NatsConn) RequestToSubject(sbj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	msg, err := nc.conn.Request(sbj, data, timeout)
	if err != nil {
		slog.Error("NatsConn RequestToSubject", "err", err.Error())
		return nil, err
	}
	return msg, nil
}

func (nc *NatsConn) Close() error {
	// TODO: we should apply graceful shutdown
	nc.conn.Close()
	return nil
}
