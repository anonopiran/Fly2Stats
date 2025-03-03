package notify

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

func Notify(rbtCfg *RbtConfig, updates []string) {
	if rbtCfg == nil {
		logrus.Error("no rabit config provided")
		return
	}
	if len(updates) == 0 {
		return
	}
	conn, err := amqp.Dial(rbtCfg.URL)
	if err != nil {
		logrus.WithError(err).Error("error dial rabbit")
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		logrus.WithError(err).Error("error getting rabbit channel")
		return
	}
	defer ch.Close()
	ctx := context.Background()
	data, _ := json.Marshal(updates)
	err = ch.PublishWithContext(ctx,
		rbtCfg.ExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:     "application/json",
			ContentEncoding: "UTF-8",
			Body:            data,
		})
	if err != nil {
		logrus.WithError(err).Error("error publishing to rabbit")
		return
	}
}
