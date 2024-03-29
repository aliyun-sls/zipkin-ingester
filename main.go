package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/aliyun-sls/zipkin-ingester/converter"
	"github.com/aliyun-sls/zipkin-ingester/exporter"
	"github.com/aliyun-sls/zipkin-ingester/receiver"
	"go.uber.org/zap"
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
	audit            bool
	protocol         string
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
	flag.StringVar(&protocol, "protocol", os.Getenv("PROTOCOL"), "protocol")
	flag.Parse()
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	audit = getAuditMode()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	var ingest receiver.Ingester
	var zipkinClient exporter.ZipkinDataExporter
	var err error
	run := true

	config := readConfiguration(sugar)

	if zipkinClient, err = exporter.NewSdkProducerExporter(config); err != nil {
		sugar.Errorw("Failed to connection sls backend", "exception", err)
		os.Exit(1)
	}

	if ingest, err = receiver.NewIngester(config, sugar); err != nil {
		sugar.Error("Failed to init kafka.", "exception", err)
		os.Exit(1)
	}

	defer ingest.Close()
	defer zipkinClient.Close()

	converter := converter.NewConverter(config.Protocol)
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			data, e := ingest.IngestTrace(sugar)

			if e == nil && audit {
				if spans, e1 := converter.ParseSpans(data, false); e1 != nil {
					sugar.Warnw("Failed to parse spans ", "Exception", e1, "originData", hex.EncodeToString(data))
				} else {
					for _, span := range spans {
						sugar.Infow("Receive Span", "TraceID", span.TraceID, "SpanID", span.ID, "parentSpanID", span.ParentID, "name", span.Name, "originData", hex.EncodeToString(data))
					}
				}
			}

			if len(data) == 0 || e != nil {
				continue
			}

			if err := zipkinClient.SendZipkinData(converter, data); err != nil {
				sugar.Warnw("Failed to send zipking data", "Exception", err, "data", hex.EncodeToString(data))
			}
		}
	}
}

func getAuditMode() bool {
	if auditMode, err := strconv.ParseBool(os.Getenv("AUDIT_MODE")); err != nil {
		return false
	} else {
		return auditMode
	}
}
func readConfiguration(sugared *zap.SugaredLogger) *configure.Configuration {
	config := &configure.Configuration{
		BootstrapServers: bootstrapServers,
		GroupID: func(consumerGroup string) string {
			if consumerGroup == "" {
				return "DEFAULT_CONSUMER_GROUP"
			}
			return consumerGroup
		}(consumerGroup),
		AutoOffsetRest: "latest",
		Topic:          strings.Split(topic, ","),
		Project:        project,
		Instance:       instance,
		AccessKey:      accessKey,
		AccessSecret:   accessSecret,
		Endpoint:       endpoint,
		Protocol:       protocol,
	}

	checkParameters(sugared, config)

	sugared.Infow("Configuration:",
		"BootstrapServers", bootstrapServers,
		"Topic", config.Topic,
		"Project", config.Project,
		"Instance", config.Instance,
		"AccessKey", config.AccessKey,
		"Endpoint", config.Endpoint,
		"Protocol", config.Protocol,
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

	if config.Protocol == "" {
		config.Protocol = "protobuf"
	}
}
