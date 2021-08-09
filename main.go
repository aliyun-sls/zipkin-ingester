package main

import (
	"flag"
	"fmt"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/aliyun-sls/zipkin-ingester/consumer"
	"github.com/aliyun-sls/zipkin-ingester/exporter"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	bootstrapservers string
	consumerGroup    string
	topic            string
	project          string
	instance         string
	accesskey        string
	accesssecret     string
	endpoint         string
)

func init() {
	flag.StringVar(&project, "project", "", "The Project name")
	flag.StringVar(&instance, "instance", "", "The instance name")
	flag.StringVar(&accesskey, "access_key", "", "The access key")
	flag.StringVar(&accesssecret, "access_secret", "", "The access secret")
	flag.StringVar(&endpoint, "endpoint", "", "The endpoint")
	flag.StringVar(&bootstrapservers, "kafka_bootstrap_services", "", "The bootstrap services")
	flag.StringVar(&consumerGroup, "kafka_consumer_group", "", "The consumer group")
	flag.StringVar(&topic, "kafka_topic", "", "The kafka topic")
	flag.Parse()
}

func main() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	var ingest consumer.Ingester
	var err error
	run := true

	config := readConfiguration()
	zipkinClient := exporter.NewZipkinExporter(config)
	if ingest, err = consumer.NewIngester(config); err != nil {
		fmt.Printf("Failed to init kafka.\n %v", err)
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
				zipkinClient.SendData(data)
			}
		}
	}
}

func readConfiguration() *configure.Configuration {

	config := &configure.Configuration{
		BootstrapServers: bootstrapservers,
		AutoOffsetRest:   "newest",
		Topic:            strings.Split(topic, ","),
		Project:          project,
		Instance:         instance,
		AccessKey:        accesskey,
		AccessSecret:     accesssecret,
		Endpoint:         endpoint,
		GroupID: func(consumerGroup string) string {
			if consumerGroup == "" {
				return "DEFAULT_CONSUMER_GROUP"
			}
			return consumerGroup
		}(consumerGroup),
	}
	return config
}
