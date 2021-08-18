package exporter

import (
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"go.uber.org/zap"
)

type ZipkinDataExporter interface {
	SendData(data []*zipkinmodel.SpanModel, sugar *zap.SugaredLogger) error
}

