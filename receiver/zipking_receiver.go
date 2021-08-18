package receiver

import "go.uber.org/zap"

type Ingester interface {
	IngestTrace(*zap.SugaredLogger) ([]byte, error)
	Close()
}
