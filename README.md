# Zipkin Ingester

Zipkin ingester is a service that reads from Kafka topic and write to SLS

## Start

1. Compile project

```shell
cd /path/to/zipkin-ingester
go build .
```

2. Start Ingester

```shell
export PROJECT=<YOUR_PROJECT>
export INSTANCE=<YOUR_INSTANCE>
export ENDPOINT=<YOUR_ENDPOINT>
export ACCESS_KEY=<YOUR_ACCESS_KEY>
export ACCESS_SECRET=<YOUR_ACCESS_SECRET>
export BOOTSTRAP_SERVICE=<YOUR_BOOTSTRAP_SERVICE>
export CONSUMER_GROUP=<YOUR_CONSUMER_GROUP>
export TOPIC=<YOUR_TOPIC>

./zipkin-ingester -project ${PROJECT} -instance ${INSTANCE} -access_key ${ACCESS_KEY} \
-access_secret ${ACCESS_SECRET} -endpoint ${endpoint} -kafka_bootstrap_services ${BOOTSTRAP_SERVICE} \
-kafka_consumer_group ${CONSUMER_GROUP} -kafka_topic ${TOPIC}
```

Have fine! :heart:


