package consumer

import (
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type Ingester interface {
	IngestTrace(*zap.SugaredLogger) ([]byte, error)
	Close()
}

type ingesterImpl struct {
	consumer *kafka.Consumer
}

func (i ingesterImpl) Close() {
	i.consumer.Close()
}

func NewIngester(config *configure.Configuration, sugar *zap.SugaredLogger) (Ingester, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  config.BootstrapServers,
		"group.id":           config.GroupID,
		"session.timeout.ms": 6000,
		"auto.offset.reset":  config.AutoOffsetRest,
	})

	if err != nil {
		sugar.Warn("Failed to new kafka consumer.", "exception", err)
		return nil, err
	}

	if e := c.SubscribeTopics(config.Topic, nil); e != nil {
		sugar.Warn("Failed to subscribe topic.", "exception", e)
		return nil, e
	} else {
		return &ingesterImpl{consumer: c}, nil
	}
}

func (i ingesterImpl) IngestTrace(suager *zap.SugaredLogger) ([]byte, error) {
	ev := i.consumer.Poll(100)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		suager.Info("Received zipkin data.", "data length", len(e.Value))
		return e.Value, nil
	case kafka.Error:
		suager.Warn("Receive a kafka error.", "Kafka error code", e.Code(), "exception", e)
		if e.Code() == kafka.ErrAllBrokersDown {
			// TODO retry
		}
		return nil, e
	default:
		return nil, nil
	}
}
