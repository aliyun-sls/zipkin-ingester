package exporter

import (
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

type ZipkinDataExporter interface {
	SendData(data []*zipkinmodel.SpanModel) error

	SendOtelData(data []*tracepb.ResourceSpans) error

	SendZipkinData(data []byte) error

	Close()
}
