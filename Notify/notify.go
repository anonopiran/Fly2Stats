package notify

import (
	config "Fly2Stats/Config"
	"context"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	amqp "github.com/rabbitmq/amqp091-go"
)

var cfg *config.SettingsType

func Notify(updates []string) error {
	if cfg.RabbitUrl == "" {
		return nil
	}
	if len(updates) == 0 {
		return nil
	}
	conn, err := amqp.Dial(cfg.RabbitUrl.AsString())
	if err != nil {
		log.Error(err)
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Error(err)
		return err
	}
	defer ch.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, _ := json.Marshal(updates)
	err = ch.PublishWithContext(ctx,
		cfg.RabbitExchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:     "application/json",
			ContentEncoding: "UTF-8",
			Body:            data,
		})
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func init() {
	cfg = config.Config()
	if cfg.RabbitUrl == "" {
		return
	}
	conn, err := amqp.Dial(cfg.RabbitUrl.AsString())
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Panic(err)

	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		cfg.RabbitExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
}
