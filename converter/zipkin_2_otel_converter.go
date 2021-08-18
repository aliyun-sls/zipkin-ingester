package converter

import (
	"encoding/json"
	"fmt"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	v11 "go.opentelemetry.io/proto/otlp/common/v1"
	v1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"math"
	"sort"
	"strconv"
	"strings"
)

type byOTLPTypes []*zipkinmodel.SpanModel

func (b byOTLPTypes) Len() int {
	return len(b)
}

func (b byOTLPTypes) Less(i, j int) bool {
	diff := strings.Compare(extractLocalServiceName(b[i]), extractLocalServiceName(b[j]))
	if diff != 0 {
		return diff <= 0
	}
	diff = strings.Compare(extractInstrumentationLibrary(b[i]), extractInstrumentationLibrary(b[j]))
	return diff <= 0
}

func (b byOTLPTypes) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func Convert2OtelSpan(spans []*zipkinmodel.SpanModel) (data []*tracepb.ResourceSpans, e error) {
	if len(spans) == 0 {
		return nil, nil
	}

	trace := make([]*tracepb.ResourceSpans, 0)

	sort.Sort(byOTLPTypes(spans))
	prevServiceName := ""
	prevInstrLibName := ""

	var curRscSpans *tracepb.ResourceSpans
	var curlIlSpans *tracepb.InstrumentationLibrarySpans
	count := 0
	for _, zspan := range spans {
		if zspan == nil {
			continue
		}

		tags := copySpanTags(zspan.Tags)
		localServiceName := extractLocalServiceName(zspan)
		if localServiceName != prevServiceName {
			prevServiceName = localServiceName
			curRscSpans = populateResourceFromZipkinSpan(tags, localServiceName)
			trace = append(trace, curRscSpans)
			count = 0
		}

		instrLibName := extractInstrumentationLibrary(zspan)
		if instrLibName != prevInstrLibName || count == 0 {
			prevInstrLibName = instrLibName
			curlIlSpans = &tracepb.InstrumentationLibrarySpans{
				Spans: make([]*tracepb.Span, 0),
			}
			curRscSpans.InstrumentationLibrarySpans = append(curRscSpans.InstrumentationLibrarySpans, curlIlSpans)
			curlIlSpans.InstrumentationLibrary = populateILFromZipkinSpan(tags, instrLibName)
			count++
		}
		if curSpan, err := zSpanToInternal(zspan, tags, true); err != nil {
			if err != nil {
				return trace, err
			}
		} else {
			if curSpan != nil {
				curlIlSpans.Spans = append(curlIlSpans.Spans, curSpan)
			}
		}
	}

	return trace, nil
}

func zSpanToInternal(zspan *zipkinmodel.SpanModel, tags map[string]string, parseStringTags bool) (*tracepb.Span, error) {
	dest := &tracepb.Span{}

	dest.TraceId = UInt64ToTraceID(zspan.TraceID.High, zspan.TraceID.Low)
	dest.SpanId = UInt64ToSpanID(uint64(zspan.ID))
	if value, ok := tags[TagW3CTraceState]; ok {
		dest.TraceState = value
		delete(tags, TagW3CTraceState)
	}
	parentID := zspan.ParentID
	if parentID != nil && *parentID != zspan.ID {
		dest.ParentSpanId = UInt64ToSpanID(uint64(*parentID))
	}

	dest.Name = zspan.Name
	dest.StartTimeUnixNano = TimestampFromTime(zspan.Timestamp)
	dest.EndTimeUnixNano = TimestampFromTime(zspan.Timestamp.Add(zspan.Duration))
	dest.Kind = zipkinKindToSpanKind(zspan.Kind, tags)

	dest.Status = populateSpanStatus(tags)
	if link, err := zTagsToSpanLinks(tags); err != nil {
		return dest, err
	} else {
		dest.Links = link
	}

	dest.Attributes = zTagsToInternalAttrs(zspan, tags, parseStringTags)
	if events, err := populateSpanEvents(zspan); err != nil {
		return nil, err
	} else {
		dest.Events = events
	}

	return dest, nil
}

func populateSpanStatus(tags map[string]string) *tracepb.Status {
	status := &tracepb.Status{}
	if value, ok := tags[TagStatusCode]; ok {
		status.Code = tracepb.Status_StatusCode(Status_StatusCode_value[value])
		delete(tags, TagStatusCode)
		if value, ok := tags[TagStatusMsg]; ok {
			status.Message = value
			delete(tags, TagStatusMsg)
		}
	}

	if val, ok := tags[TagError]; ok {
		if val == "true" {
			status.Code = tracepb.Status_STATUS_CODE_ERROR
			delete(tags, TagError)
		}
	}

	return status
}

func zipkinKindToSpanKind(kind zipkinmodel.Kind, tags map[string]string) tracepb.Span_SpanKind {
	switch kind {
	case zipkinmodel.Client:
		return tracepb.Span_SPAN_KIND_CLIENT
	case zipkinmodel.Server:
		return tracepb.Span_SPAN_KIND_SERVER
	case zipkinmodel.Producer:
		return tracepb.Span_SPAN_KIND_PRODUCER
	case zipkinmodel.Consumer:
		return tracepb.Span_SPAN_KIND_CONSUMER
	default:
		if value, ok := tags[TagSpanKind]; ok {
			delete(tags, TagSpanKind)
			if value == "internal" {
				return tracepb.Span_SPAN_KIND_INTERNAL
			}
		}
		return tracepb.Span_SPAN_KIND_UNSPECIFIED
	}
}

func zTagsToSpanLinks(tags map[string]string) ([]*tracepb.Span_Link, error) {
	var links []*tracepb.Span_Link
	for i := 0; i < 128; i++ {
		key := fmt.Sprintf("otlp.link.%d", i)
		val, ok := tags[key]
		if !ok {
			return nil, nil
		}
		delete(tags, key)

		parts := strings.Split(val, "|")
		partCnt := len(parts)
		if partCnt < 5 {
			continue
		}

		link := &tracepb.Span_Link{}
		links = append(links, link)

		if rawTrace, errTrace := UnmarshalTraceJSON([]byte(parts[0])); errTrace != nil {
			return nil, errTrace
		} else {
			link.TraceId = rawTrace
		}

		if rawSpan, errSpan := UnmarshalSpanJSON([]byte(parts[1])); errSpan != nil {
			return nil, errSpan
		} else {
			link.SpanId = rawSpan
		}

		link.TraceState = parts[2]
		var jsonStr string
		if partCnt == 5 {
			jsonStr = parts[3]
		} else {
			jsonParts := parts[3 : partCnt-1]
			jsonStr = strings.Join(jsonParts, "|")
		}
		var attrs map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &attrs); err != nil {
			return nil, err
		}
		if data, err := jsonMapToAttributeMap(attrs); err != nil {
			return nil, err
		} else {
			link.Attributes = data
		}

		dropped, errDropped := strconv.ParseUint(parts[partCnt-1], 10, 32)
		if errDropped != nil {
			return nil, errDropped
		}
		link.DroppedAttributesCount = uint32(dropped)
	}
	return nil, nil
}

func populateSpanEvents(zspan *zipkinmodel.SpanModel) (data []*tracepb.Span_Event, e error) {
	data = make([]*tracepb.Span_Event, len(zspan.Annotations))
	for _, anno := range zspan.Annotations {
		event := &tracepb.Span_Event{}
		event.TimeUnixNano = TimestampFromTime(anno.Timestamp)

		parts := strings.Split(anno.Value, "|")
		partCnt := len(parts)
		event.Name = parts[0]
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
		}
		if attrs, err := jsonMapToAttributeMap(attrs); err != nil {
			return nil, err
		} else {
			event.Attributes = attrs
		}

		dropped, errDropped := strconv.ParseUint(parts[partCnt-1], 10, 32)
		if errDropped != nil {
			return nil, errDropped
		}
		event.DroppedAttributesCount = uint32(dropped)

		data = append(data, event)
	}
	return data, nil
}

func jsonMapToAttributeMap(attrs map[string]interface{}) (data []*v11.KeyValue, e error) {
	for key, val := range attrs {
		attr := &v11.KeyValue{}
		attr.Key = key
		if s, ok := val.(string); ok {
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: s}}
		} else if d, ok := val.(float64); ok {
			if math.Mod(d, 1.0) == 0.0 {
				attr.Value = &v11.AnyValue{Value: &v11.AnyValue_IntValue{IntValue: int64(d)}}
			} else {
				attr.Value = &v11.AnyValue{Value: &v11.AnyValue_DoubleValue{DoubleValue: d}}
			}
		} else if b, ok := val.(bool); ok {
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_BoolValue{BoolValue: b}}
		}

	}
	return data, nil
}

func zTagsToInternalAttrs(zspan *zipkinmodel.SpanModel, tags map[string]string, parseStringTags bool) (data []*v11.KeyValue) {
	data = tagsToAttributeMap(tags, parseStringTags)
	if zspan.LocalEndpoint != nil {
		if zspan.LocalEndpoint.IPv4 != nil {
			attr := &v11.KeyValue{}
			attr.Key = AttributeNetHostIP
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: zspan.LocalEndpoint.IPv4.String()}}
			data = append(data, attr)
		}
		if zspan.LocalEndpoint.IPv6 != nil {
			attr := &v11.KeyValue{}
			attr.Key = AttributeNetHostIP
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: zspan.LocalEndpoint.IPv6.String()}}
			data = append(data, attr)
		}
		if zspan.LocalEndpoint.Port > 0 {
			attr := &v11.KeyValue{}
			attr.Key = AttributeNetHostPort
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_IntValue{IntValue: int64(zspan.LocalEndpoint.Port)}}
			data = append(data, attr)
		}
	}
	if zspan.RemoteEndpoint != nil {
		attr := &v11.KeyValue{}
		attr.Key = AttributeNetHostIP

		if zspan.RemoteEndpoint.ServiceName != "" {
			attr := &v11.KeyValue{}
			attr.Key = AttributePeerService
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: zspan.RemoteEndpoint.ServiceName}}
			data = append(data, attr)
		}
		if zspan.RemoteEndpoint.IPv4 != nil {
			attr := &v11.KeyValue{}
			attr.Key = AttributeNetPeerIP
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: zspan.RemoteEndpoint.IPv4.String()}}
			data = append(data, attr)
		}
		if zspan.RemoteEndpoint.IPv6 != nil {
			attr := &v11.KeyValue{}
			attr.Key = AttributeNetPeerIP
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: zspan.RemoteEndpoint.IPv6.String()}}
			data = append(data, attr)
		}
		if zspan.RemoteEndpoint.Port > 0 {
			attr := &v11.KeyValue{}
			attr.Key = AttributeNetPeerPort
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_IntValue{IntValue: int64(zspan.RemoteEndpoint.Port)}}
			data = append(data, attr)
		}
	}
	return data
}

func tagsToAttributeMap(tags map[string]string, parseStringTags bool) (data []*v11.KeyValue) {
	for key, val := range tags {
		if _, ok := nonSpanAttributes[key]; ok {
			continue
		}

		d := &v11.KeyValue{}
		d.Key = key
		if parseStringTags {
			switch DetermineValueType(val, false) {
			case AttributeValueINT:
				iValue, _ := strconv.ParseInt(val, 10, 64)
				d.Value = &v11.AnyValue{Value: &v11.AnyValue_IntValue{IntValue: iValue}}
			case AttributeValueDOUBLE:
				fValue, _ := strconv.ParseFloat(val, 64)
				d.Value = &v11.AnyValue{Value: &v11.AnyValue_DoubleValue{DoubleValue: fValue}}
			case AttributeValueBOOL:
				bValue, _ := strconv.ParseBool(val)
				d.Value = &v11.AnyValue{Value: &v11.AnyValue_BoolValue{BoolValue: bValue}}
			default:
				d.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: val}}
			}
		} else {
			d.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: val}}
		}

		data = append(data, d)
	}
	return data
}

func populateResourceFromZipkinSpan(tags map[string]string, localServiceName string) (data *tracepb.ResourceSpans) {
	if localServiceName == ResourceNoServiceName {
		return nil
	}

	data = &tracepb.ResourceSpans{
		Resource: &v1.Resource{},
	}

	if len(tags) == 0 {
		attr := &v11.KeyValue{}
		attr.Key = AttributeServiceName
		attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: localServiceName}}
		data.Resource.Attributes = append(data.Resource.Attributes, attr)
		return
	}

	snSource := tags[TagServiceNameSource]
	if snSource == "" {
		attr := &v11.KeyValue{}
		attr.Key = AttributeServiceName
		attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: localServiceName}}
		data.Resource.Attributes = append(data.Resource.Attributes, attr)
	} else {
		attr := &v11.KeyValue{}
		attr.Key = snSource
		attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: localServiceName}}
		data.Resource.Attributes = append(data.Resource.Attributes, attr)
	}
	delete(tags, TagServiceNameSource)

	for key := range getNonSpanAttributes() {
		if key == TagInstrumentationName || key == TagInstrumentationVersion {
			continue
		}
		if value, ok := tags[key]; ok {
			attr := &v11.KeyValue{}
			attr.Key = key
			attr.Value = &v11.AnyValue{Value: &v11.AnyValue_StringValue{StringValue: value}}
			data.Resource.Attributes = append(data.Resource.Attributes, attr)
			delete(tags, key)
		}
	}

	return data
}

func populateILFromZipkinSpan(tags map[string]string, instrLibName string) (data *v11.InstrumentationLibrary) {
	if instrLibName == "" {
		return nil
	}
	data = &v11.InstrumentationLibrary{}
	if value, ok := tags[TagInstrumentationName]; ok {
		data.Name = value
		delete(tags, TagInstrumentationName)
	}
	if value, ok := tags[TagInstrumentationVersion]; ok {
		data.Version = value
		delete(tags, TagInstrumentationVersion)
	}
	return data
}

func copySpanTags(tags map[string]string) map[string]string {
	dest := make(map[string]string, len(tags))
	for key, val := range tags {
		dest[key] = val
	}
	return dest
}

func extractLocalServiceName(zspan *zipkinmodel.SpanModel) string {
	if zspan == nil || zspan.LocalEndpoint == nil || zspan.LocalEndpoint.ServiceName == "" {
		return ResourceNoServiceName
	}
	return zspan.LocalEndpoint.ServiceName
}

func extractInstrumentationLibrary(zspan *zipkinmodel.SpanModel) string {
	if zspan == nil || len(zspan.Tags) == 0 {
		return ""
	}
	return zspan.Tags[TagInstrumentationName]
}
