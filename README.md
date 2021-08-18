# Zipkin Ingester

Zipkin Ingester支持从Kafka消费Zipkin的数据，支持协议：proto协议

## Start

1. 编译工程

```shell
cd /path/to/zipkin-ingester
go build .
```

2. 启动Ingester

```shell
export PROJECT=<YOUR_PROJECT>
export INSTANCE=<YOUR_INSTANCE>
export ENDPOINT=<YOUR_ENDPOINT>
export ACCESS_KEY=<YOUR_ACCESS_KEY>
export ACCESS_SECRET=<YOUR_ACCESS_SECRET>
export BOOTSTRAP_SERVICE=<YOUR_BOOTSTRAP_SERVICE>
export CONSUMER_GROUP=<YOUR_CONSUMER_GROUP>
export TOPIC=<YOUR_TOPIC>
export AUDIT_MODE=false

./zipkin-ingester -project ${PROJECT} -instance ${INSTANCE} -access_key ${ACCESS_KEY} \
-access_secret ${ACCESS_SECRET} -endpoint ${ENDPOINT} -kafka_bootstrap_services ${BOOTSTRAP_SERVICE} \
-kafka_consumer_group ${CONSUMER_GROUP} -kafka_topic ${TOPIC}
```

各参数详细介绍:

|参数|描述|
|:---|:---|
|ACCESS_KEY_ID| 阿里云账号AccessKey ID。<br/>建议您使用只具备日志服务Project写入权限的RAM用户的AccessKey（包括AccessKey ID和AccessKey Secret）。|
|ACCESS_SECRET| 阿里云账号AccessKey Secret。<br/>建议您使用只具备日志服务Project写入权限的RAM用户的AccessKey。|
|PROJECT_NAME|日志服务Project名称。 |
|INSTANCE|Trace服务实例名称。 |
|ENDPOINT|接入地址，格式为${project}.${region-endpoint}:10010，其中：<br/> ${project}：日志服务Project名称。<br/>${region-endpoint}：Project访问域名，支持公网和阿里云内网（经典网络、VPC）。 |
|BOOTSTRAP_SERVICE|Kafka服务地址。 |
|CONSUMER_GROUP|kafka消费组。 |
|TOPIC| Kafka Topic |

Have fine! :heart:


