package exporter

import (
	"context"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"github.com/aliyun-sls/zipkin-ingester/converter"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc/credentials"
)

type grpcOtelDataExporter struct {
	client otlptrace.Client
}

func (g *grpcOtelDataExporter) SendZipkinData(data []byte) error {
	if spans, e1 := converter.ParseSpans(data, false); e1 == nil {
		return g.SendData(spans)
	} else {
		return e1
	}
}

func NewGrpcOtelDataExporter(configure *configure.Configuration) (ZipkinDataExporter, error) {
	headers := make(map[string]string)
	headers["x-sls-otel-project"] = configure.Project
	headers["x-sls-otel-instance-id"] = configure.Instance
	headers["x-sls-otel-ak-id"] = configure.AccessKey
	headers["x-sls-otel-ak-secret"] = configure.AccessSecret

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlptracegrpc.WithEndpoint(configure.Endpoint),
		otlptracegrpc.WithHeaders(headers),
	)

	if err := client.Start(context.Background()); err != nil {
		return nil, err
	}

	return &grpcOtelDataExporter{
		client: client,
	}, nil
}

func (g *grpcOtelDataExporter) SendData(data []*zipkinmodel.SpanModel) error {
	if spans, err := converter.Convert2OtelSpan(data); err == nil {
		return g.SendOtelData(spans)
	} else {
		return err
	}
}

func (g *grpcOtelDataExporter) SendOtelData(data []*tracepb.ResourceSpans) error {
	return g.client.UploadTraces(context.Background(), data)
}
