module github.com/aliyun-sls/zipkin-ingester

go 1.16

replace github.com/aliyun-sls/zipkin-ingester => ./

require (
	github.com/aliyun/aliyun-log-go-sdk v0.1.21
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/viper v1.8.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.0.0-RC2
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC2
	go.opentelemetry.io/proto/otlp v0.9.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.0
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
)
