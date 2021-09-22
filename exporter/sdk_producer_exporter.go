package exporter

import (
	"fmt"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/aliyun-sls/zipkin-ingester/converter"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type SdkProducerExporter struct {
	project          string
	traceLog         string
	producerInstance *producer.Producer
}

func (s *SdkProducerExporter) Close() {
	s.producerInstance.Close(60 * 1000)
}

func NewSdkProducerExporter(configure *configure.Configuration) (ZipkinDataExporter, error) {
	producerConfig := producer.GetDefaultProducerConfig()
	producerConfig.Endpoint = configure.Endpoint
	producerConfig.AccessKeyID = configure.AccessKey
	producerConfig.AccessKeySecret = configure.AccessSecret
	producerInstance := producer.InitProducer(producerConfig)
	producerInstance.Start()
	return &SdkProducerExporter{
		producerInstance: producerInstance,
		project:          configure.Project,
		traceLog:         fmt.Sprintf("%s-traces", configure.Instance),
	}, nil
}

func (s SdkProducerExporter) SendData(data []*zipkinmodel.SpanModel) error {
	converter.SendToSls(data, s.producerInstance, s.project, s.traceLog)
	return nil
}

func (s SdkProducerExporter) SendOtelData(data []*tracepb.ResourceSpans) error {
	panic("Unsuppport")
}

func (s SdkProducerExporter) SendZipkinData(data []byte) error {
	if spans, err := converter.ParseSpans(data, false); err == nil {
		return s.SendData(spans)
	} else {
		return err
	}
}
