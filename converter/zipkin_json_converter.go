package converter

import (
	"encoding/json"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
)

type JsonConvertor struct {
}

type zipkinJsonModel struct {
	TraceID   string
	ID        string
	Kind      string
	Name      string
	Timestamp uint64
	Duration  uint64
	ParentID  string
}

func (c *JsonConvertor) ParseSpans(protoBlob []byte, debugWasSet bool) (zss []*zipkinmodel.SpanModel, err error) {
	var models []map[string]interface{}
	if e := json.Unmarshal(protoBlob, &models); e != nil {
		return nil, e
	}

	for _, model := range models {
		if zms, e := c.parseSpan(model); e != nil {
			return nil, e
		} else {
			zss = append(zss, zms)
		}
	}

	return zss, nil
}

func (c *JsonConvertor) parseSpan(model map[string]interface{}) (spanModel *zipkinmodel.SpanModel, e error) {
	var data []byte
	if data, e = json.Marshal(model); e != nil {
		return nil, e
	}

	spanmodel := &zipkinmodel.SpanModel{}
	if e = spanmodel.UnmarshalJSON(data); e != nil {
		return nil, e
	} else {
		return spanmodel, nil
	}
}
