package exporter

import (
	"fmt"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/aliyun-sls/zipkin-ingester/converter"
	slsSdk "github.com/aliyun/aliyun-log-go-sdk"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type SdkDataExporter struct {
	client   *slsSdk.Client
	project  string
	traceLog string
}

func (s *SdkDataExporter) Close() {
	s.client.Close()
}

func NewSdkDataExporter(configure *configure.Configuration) (ZipkinDataExporter, error) {
	return &SdkDataExporter{
		client: &slsSdk.Client{
			Endpoint:        configure.Endpoint,
			AccessKeyID:     configure.AccessKey,
			AccessKeySecret: configure.AccessSecret,
		},
		project:  configure.Project,
		traceLog: fmt.Sprintf("%s-traces", configure.Instance),
	}, nil
}

func (s SdkDataExporter) SendData(data []*zipkinmodel.SpanModel) error {
	if lg, err := converter.ToSLSSpans(data); err != nil {
		return err
	} else {
		return s.client.PutLogs(s.project, s.traceLog, lg)
	}
}

func (s SdkDataExporter) SendOtelData(data []*tracepb.ResourceSpans) error {
	panic("Unsupported")
}

func (s SdkDataExporter) SendZipkinData(converter converter.Converter, data []byte) error {
	if spans, err := converter.ParseSpans(data, false); err == nil {
		return s.SendData(spans)
	} else {
		return err
	}
}
