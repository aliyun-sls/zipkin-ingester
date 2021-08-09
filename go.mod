module github.com/aliyun-sls/zipkin-ingester

go 1.16

replace github.com/aliyun-sls/zipkin-ingester => ./

require (
	github.com/confluentinc/confluent-kafka-go v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1 // indirect
)
