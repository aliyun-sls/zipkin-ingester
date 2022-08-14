package converter

import (
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"strings"
)

type Converter interface {
	ParseSpans(protoBlob []byte, debugWasSet bool) (zss []*zipkinmodel.SpanModel, err error)
}

func NewConverter(protocol string) Converter {
	if strings.ToUpper(protocol) == "JSON" {
		return &JsonConvertor{}
	} else {
		return &ProtobufConvertor{}
	}
}
