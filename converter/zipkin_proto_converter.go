package converter

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/openzipkin/zipkin-go/proto/zipkin_proto3"
	"net"
	"time"

	"google.golang.org/protobuf/proto"

	zipkinmodel "github.com/openzipkin/zipkin-go/model"
)

// Copy from zipking-go
func ParseSpans(protoBlob []byte, debugWasSet bool) (zss []*zipkinmodel.SpanModel, err error) {
	var listOfSpans zipkin_proto3.ListOfSpans
	if err := proto.Unmarshal(protoBlob, &listOfSpans); err != nil {
		return nil, err
	}
	for _, zps := range listOfSpans.Spans {
		zms, err := protoSpanToModelSpan(zps, debugWasSet)
		if err != nil {
			fmt.Printf("Failed to convert span, %s\n", err.Error())
			continue
		}
		zss = append(zss, zms)
	}
	return zss, nil
}

var errNilZipkinSpan = errors.New("expecting a non-nil Span")

func protoSpanToModelSpan(s *zipkin_proto3.Span, debugWasSet bool) (*zipkinmodel.SpanModel, error) {
	if s == nil {
		return nil, errNilZipkinSpan
	}
	traceID, err := zipkinmodel.TraceIDFromHex(fmt.Sprintf("%x", s.TraceId))
	if err != nil {
		return nil, fmt.Errorf("invalid TraceID: %v", err)
	}

	parentSpanID, _, err := protoSpanIDToModelSpanID(s.ParentId)
	if err != nil {
		return nil, fmt.Errorf("invalid ParentID: %v", err)
	}
	spanIDPtr, spanIDBlank, err := protoSpanIDToModelSpanID(s.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid SpanID: %v", err)
	}
	if spanIDBlank || spanIDPtr == nil {
		// This is a logical error
		return nil, errors.New("expected a non-nil SpanID")
	}

	zmsc := zipkinmodel.SpanContext{
		TraceID:  traceID,
		ID:       *spanIDPtr,
		ParentID: parentSpanID,
		Debug:    debugWasSet,
	}
	zms := &zipkinmodel.SpanModel{
		SpanContext:    zmsc,
		Name:           s.Name,
		Kind:           zipkinmodel.Kind(s.Kind.String()),
		Timestamp:      microsToTime(s.Timestamp),
		Tags:           s.Tags,
		Duration:       microsToDuration(s.Duration),
		LocalEndpoint:  protoEndpointToModelEndpoint(s.LocalEndpoint),
		RemoteEndpoint: protoEndpointToModelEndpoint(s.RemoteEndpoint),
		Shared:         s.Shared,
		Annotations:    protoAnnotationsToModelAnnotations(s.Annotations),
	}

	if uint32(zms.Timestamp.Unix()) <= 0 {
		return nil, errors.New("StartTime is zero")
	}

	return zms, nil
}

func microsToDuration(us uint64) time.Duration {
	return time.Duration(us * 1e3)
}

func protoEndpointToModelEndpoint(zpe *zipkin_proto3.Endpoint) *zipkinmodel.Endpoint {
	if zpe == nil {
		return nil
	}
	return &zipkinmodel.Endpoint{
		ServiceName: zpe.ServiceName,
		IPv4:        net.IP(zpe.Ipv4),
		IPv6:        net.IP(zpe.Ipv6),
		Port:        uint16(zpe.Port),
	}
}

func protoSpanIDToModelSpanID(spanId []byte) (zid *zipkinmodel.ID, blank bool, err error) {
	if len(spanId) == 0 {
		return nil, true, nil
	}
	if len(spanId) != 8 {
		return nil, true, fmt.Errorf("has length %d yet wanted length 8", len(spanId))
	}

	u64 := binary.BigEndian.Uint64(spanId)
	zid_ := zipkinmodel.ID(u64)
	return &zid_, false, nil
}

func protoAnnotationsToModelAnnotations(zpa []*zipkin_proto3.Annotation) (zma []zipkinmodel.Annotation) {
	for _, za := range zpa {
		if za != nil {
			zma = append(zma, zipkinmodel.Annotation{
				Timestamp: microsToTime(za.Timestamp),
				Value:     za.Value,
			})
		}
	}

	if len(zma) == 0 {
		return nil
	}
	return zma
}

func microsToTime(us uint64) time.Time {
	return time.Unix(0, int64(us*1e3)).UTC()
}
