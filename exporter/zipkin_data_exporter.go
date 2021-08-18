package exporter

import (
	"go.uber.org/zap"
)

type ZipkinDataExporter interface {
	SendData(data []byte, sugar *zap.SugaredLogger) error
}

