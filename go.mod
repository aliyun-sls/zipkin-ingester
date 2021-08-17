module github.com/aliyun-sls/zipkin-ingester

go 1.16

replace github.com/aliyun-sls/zipkin-ingester => ./

require (
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/openzipkin/zipkin-go v0.2.5 // indirect
	github.com/spf13/viper v1.8.1
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.0
)
