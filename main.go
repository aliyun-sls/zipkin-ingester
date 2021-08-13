package main

import (
	"flag"
	"fmt"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/aliyun-sls/zipkin-ingester/consumer"
	"github.com/aliyun-sls/zipkin-ingester/exporter"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	bootstrapServers string
	consumerGroup    string
	topic            string
	project          string
	instance         string
	accessKey        string
	accessSecret     string
	endpoint         string
)

func init() {
	flag.StringVar(&project, "project", os.Getenv("PROJECT"), "The Project name")
	flag.StringVar(&instance, "instance", os.Getenv("INSTANCE"), "The instance name")
	flag.StringVar(&accessKey, "access_key", os.Getenv("ACCESS_KEY"), "The access key")
	flag.StringVar(&accessSecret, "access_secret", os.Getenv("ACCESS_SECRET"), "The access secret")
	flag.StringVar(&endpoint, "endpoint", os.Getenv("ENDPOINT"), "The endpoint")
	flag.StringVar(&bootstrapServers, "kafka_bootstrap_services", os.Getenv("BOOTSTRAP_SERVICE"), "The bootstrap services")
	flag.StringVar(&consumerGroup, "kafka_consumer_group", os.Getenv("CONSUMER_GROUP"), "The consumer group")
	flag.StringVar(&topic, "kafka_topic", os.Getenv("TOPIC"), "The kafka topic")
	flag.Parse()
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	var ingest consumer.Ingester
	var err error
	run := true

	config := readConfiguration(sugar)
	zipkinClient := exporter.NewZipkinExporter(config)
	if ingest, err = consumer.NewIngester(config, sugar); err != nil {
		sugar.Warn("Failed to init kafka.", "exception", err)
		os.Exit(1)
	} else {
		defer ingest.Close()
	}

	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			data, err := ingest.IngestTrace()
			if err == nil && data != nil {
				zipkinClient.SendData(data, sugar)
			}
		}
	}
}

func readConfiguration(sugared *zap.SugaredLogger) *configure.Configuration {
	config := &configure.Configuration{
		BootstrapServers: bootstrapServers,
		AutoOffsetRest:   "latest",
		Topic:            strings.Split(topic, ","),
		Project:          project,
		Instance:         instance,
		AccessKey:        accessKey,
		AccessSecret:     accessSecret,
		Endpoint:         endpoint,
		GroupID: func(consumerGroup string) string {
			if consumerGroup == "" {
				return "DEFAULT_CONSUMER_GROUP"
			}
			return consumerGroup
		}(consumerGroup),
	}

	checkParameters(sugared, config)

	sugared.Info("Configuration:",
		"BootstrapServers", bootstrapServers,
		"Topic", config.Topic,
		"Project", config.Project,
		"Instance", config.Instance,
		"AccessKey", config.AccessKey,
		"Endpoint", config.Endpoint,
	)
	return config
}

func checkParameters(sugared *zap.SugaredLogger, config *configure.Configuration) {
	if config.BootstrapServers == "" {
		sugared.Warn("The bootstrap servers is empty.")
		panic("The bootstrap servers is empty")
	}

	if len(config.Topic) == 0 {
		sugared.Warn("The topic is empty.")
		panic("The topic is empty.")
	}

	if config.Project == "" {
		sugared.Warn("The project is empty.")
		panic("The project is empty.")
	}

	if config.Instance == "" {
		sugared.Warn("The instance is empty.")
		panic("The instance is empty.")
	}

	if config.AccessKey == "" {
		sugared.Warn("The access key is empty.")
		panic("The access key is empty")
	}

	if config.AccessSecret == "" {
		sugared.Warn("The access secret is empty.")
		panic("The access secret is empty")
	}

	if config.Endpoint == "" {
		sugared.Warn("The endpoint is empty.")
		panic("The endpoint is empty")
	}
}
