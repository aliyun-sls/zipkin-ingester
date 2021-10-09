package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	slsSdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/gogo/protobuf/proto"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	"github.com/spf13/cast"
	v11 "go.opentelemetry.io/proto/otlp/common/v1"
	"strings"
)

func ToSLSSpans(spans []*zipkinmodel.SpanModel) (lg *slsSdk.LogGroup, err error) {
	lg = &slsSdk.LogGroup{
		Topic:  proto.String("0.0.0.0"),
		Source: proto.String(""),
	}

	for _, span := range spans {
		if log, err := spanToLog(span); err == nil {
			lg.Logs = append(lg.Logs, log)
		} else {
			continue
		}
	}
	return lg, nil
}

func SendToSls(spans []*zipkinmodel.SpanModel, instance *producer.Producer, callback producer.CallBack, project string, log string) error {
	for _, span := range spans {
		go convertAndSend(span, instance, callback, project, log)
	}

	return nil
}

func convertAndSend(span *zipkinmodel.SpanModel, instance *producer.Producer, callback producer.CallBack, project string, traceLogstore string) {
	if log, err := spanToLog(span); err == nil {
		fmt.Printf("%v\n", log)
		error := instance.SendLogWithCallBack(project, traceLogstore, "0.0.0.0", "", log, callback)
		if error != nil {
			fmt.Printf("%v", error)
		}
	} else {
		fmt.Printf("%v, %v\n", err, span)
	}
}

func spanToLog(span *zipkinmodel.SpanModel) (*slsSdk.Log, error) {
	contents, err := ToSLSSpan(span)
	if err != nil {
		return nil, err
	}

	if uint32(span.Timestamp.Unix()) <= 0 {
		return nil, errors.New(fmt.Sprintf("time is zero. %v", span.Timestamp.Unix()))
	}

	return &slsSdk.Log{
		Time:     proto.Uint32(uint32(span.Timestamp.Unix())),
		Contents: contents,
	}, nil
}

func ToSLSSpan(span *zipkinmodel.SpanModel) ([]*slsSdk.LogContent, error) {
	contents := make([]*slsSdk.LogContent, 0)
	tags := copySpanTags(span.Tags)
	localServiceName := extractLocalServiceName(span)
	contents = appendAttributeToLogContent(contents, OperationName, span.Name)
	contents = appendAttributeToLogContent(contents, StartTime, cast.ToString(span.Timestamp.UnixNano()/1000))
	contents = appendAttributeToLogContent(contents, Duration, cast.ToString(span.Duration.Nanoseconds()/1000))
	contents = appendAttributeToLogContent(contents, EndTime, cast.ToString((span.Timestamp.UnixNano()+span.Duration.Nanoseconds())/1000))
	contents = appendAttributeToLogContent(contents, ServiceName, localServiceName)
	contents = appendAttributeToLogContent(contents, SpanKind, strings.ToLower(string(span.Kind)))
	contents = appendAttributeToLogContent(contents, TraceID, span.TraceID.String())
	contents = appendAttributeToLogContent(contents, SpanID, span.ID.String())

	if resource, err := extractResource(tags, localServiceName); err == nil {
		contents = appendAttributeToLogContent(contents, Resource, string(resource))
	}

	if tags[TagStatusCode] != "" {
		contents = appendAttributeToLogContent(contents, StatusCode, tags[TagStatusCode])
	} else {
		contents = appendAttributeToLogContent(contents, StatusCode, "UNSET")
	}

	if span.ParentID != nil {
		contents = appendAttributeToLogContent(contents, ParentSpanID, span.ParentID.String())
	} else {
		contents = appendAttributeToLogContent(contents, ParentSpanID, "")
	}

	if links, err := extractLinks(tags); err == nil {
		contents = appendAttributeToLogContent(contents, Links, string(links))
	}

	if t, err := extractTags(span, tags); err == nil {
		contents = appendAttributeToLogContent(contents, Attribute, string(t))
	}

	if logs, err := extractLogs(span); err == nil {
		contents = appendAttributeToLogContent(contents, Logs, string(logs))
	}

	return contents, nil
}

func extractResource(tags map[string]string, localServiceName string) ([]byte, error) {
	resources := make(map[string]string)
	resources[AttributeServiceName] = localServiceName

	if len(tags) == 0 {
		return json.Marshal(resources)
	}

	snSource := tags[TagServiceNameSource]
	if snSource == "" {
		resources[AttributeServiceName] = localServiceName
	} else {
		resources[snSource] = localServiceName
	}
	delete(tags, TagServiceNameSource)

	for key := range getNonSpanAttributes() {
		if key == TagInstrumentationName || key == TagInstrumentationVersion {
			continue
		}
		if value, ok := tags[key]; ok {
			resources[key] = value
			delete(tags, key)
		}
	}

	return json.Marshal(resources)
}

func extractLogs(zspan *zipkinmodel.SpanModel) ([]byte, error) {
	slsLogs := make([]*SpanLog, len(zspan.Annotations))
	for _, anno := range zspan.Annotations {
		event := &SpanLog{
			Attribute: make(map[string]interface{}),
		}
		event.Time = TimestampFromTime(anno.Timestamp)

		parts := strings.Split(anno.Value, "|")
		partCnt := len(parts)
		event.Attribute["Name"] = parts[0]
		if partCnt < 3 {
			continue
		}

		var jsonStr string
		if partCnt == 3 {
			jsonStr = parts[1]
		} else {
			jsonParts := parts[1 : partCnt-1]
			jsonStr = strings.Join(jsonParts, "|")
		}
		var attrs map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &attrs); err != nil {
			return nil, err
		} else {
			for key, value := range attrs {
				event.Attribute[key] = value
			}
		}

		slsLogs = append(slsLogs, event)
	}

	return json.Marshal(slsLogs)
}

func extractTags(zspan *zipkinmodel.SpanModel, tags map[string]string) ([]byte, error) {
	result := make(map[string]interface{})

	for key, val := range tags {
		if _, ok := nonSpanAttributes[key]; ok {
			continue
		}
		d := &v11.KeyValue{}
		d.Key = key
		result[key] = val
	}

	if zspan.LocalEndpoint != nil {
		if zspan.LocalEndpoint.IPv4 != nil {
			result[AttributeNetHostIP] = zspan.LocalEndpoint.IPv4.String()
		}
		if zspan.LocalEndpoint.IPv6 != nil {
			result[AttributeNetHostIP] = zspan.LocalEndpoint.IPv6.String()
		}
		if zspan.LocalEndpoint.Port > 0 {
			result[AttributeNetHostPort] = zspan.LocalEndpoint.Port
		}
	}
	if zspan.RemoteEndpoint != nil {
		attr := &v11.KeyValue{}
		attr.Key = AttributeNetHostIP

		if zspan.RemoteEndpoint.ServiceName != "" {
			result[AttributePeerService] = zspan.RemoteEndpoint.ServiceName
		}
		if zspan.RemoteEndpoint.IPv4 != nil {
			result[AttributeNetPeerIP] = zspan.RemoteEndpoint.IPv4.String()
		}
		if zspan.RemoteEndpoint.IPv6 != nil {
			result[AttributeNetPeerIP] = zspan.RemoteEndpoint.IPv6.String()
		}
		if zspan.RemoteEndpoint.Port > 0 {
			result[AttributeNetPeerIP] = zspan.RemoteEndpoint.Port
		}
	}
	return json.Marshal(result)
}

func extractLinks(tags map[string]string) ([]byte, error) {
	links := make([]map[string]string, 0)

	for i := 0; i < 128; i++ {
		key := fmt.Sprintf("otlp.link.%d", i)
		val, ok := tags[key]
		if !ok {
			return []byte("[]"), nil
		}
		delete(tags, key)

		parts := strings.Split(val, "|")
		partCnt := len(parts)
		if partCnt < 5 {
			continue
		}

		link := make(map[string]string)
		links = append(links, link)

		link[TraceID] = parts[0]
		link[SpanID] = parts[1]
		link["RefType"] = parts[2]
	}
	return json.Marshal(links)
}

type SpanLog struct {
	Attribute map[string]interface{} `json:"attribute"`
	Time      uint64                 `json:"time"`
}

func appendAttributeToLogContent(contents []*slsSdk.LogContent, k, v string) []*slsSdk.LogContent {
	content := slsSdk.LogContent{
		Key:   proto.String(k),
		Value: proto.String(v),
	}
	return append(contents, &content)
}
