package converter

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"time"
)

func UInt64ToTraceID(high, low uint64) []byte {
	traceID := make([]byte, 16)
	binary.BigEndian.PutUint64(traceID[:8], high)
	binary.BigEndian.PutUint64(traceID[8:], low)
	return traceID
}

func UInt64ToSpanID(id uint64) []byte {
	spanID := make([]byte, 8)
	binary.BigEndian.PutUint64(spanID[:], id)
	return spanID
}

func TimestampFromTime(t time.Time) uint64 {
	return uint64(t.UnixNano())
}

var Status_StatusCode_value = map[string]int32{
	"STATUS_CODE_UNSET": 0,
	"STATUS_CODE_OK":    1,
	"STATUS_CODE_ERROR": 2,
}

func unmarshalJSON(dst []byte, src []byte) (e error) {
	if l := len(src); l >= 2 && src[0] == '"' && src[l-1] == '"' {
		src = src[1 : l-1]
	}
	nLen := len(src)
	if nLen == 0 {
		return nil
	}

	if len(dst) != hex.DecodedLen(nLen) {
		return errors.New("invalid length for ID")
	}

	_, err := hex.Decode(dst, src)
	if err != nil {
		return fmt.Errorf("cannot unmarshal ID from string '%s': %w", string(src), err)
	}
	return nil
}

var nonSpanAttributes = getNonSpanAttributes()

func getNonSpanAttributes() map[string]struct{} {
	attrs := make(map[string]struct{})
	for _, key := range GetResourceSemanticConventionAttributeNames() {
		attrs[key] = struct{}{}
	}
	attrs[TagServiceNameSource] = struct{}{}
	attrs[TagInstrumentationName] = struct{}{}
	attrs[TagInstrumentationVersion] = struct{}{}
	attrs[OCAttributeProcessStartTime] = struct{}{}
	attrs[OCAttributeExporterVersion] = struct{}{}
	attrs[OCAttributeProcessID] = struct{}{}
	attrs[OCAttributeResourceType] = struct{}{}
	return attrs
}

var complexAttrValDescriptions = getComplexAttrValDescripts()
var attrValDescriptions = getAttrValDescripts()

type attrValDescript struct {
	regex    *regexp.Regexp
	attrType AttributeValueType
}

func getComplexAttrValDescripts() []*attrValDescript {
	descriptions := getAttrValDescripts()
	return descriptions[4:]
}

func getAttrValDescripts() []*attrValDescript {
	descriptions := make([]*attrValDescript, 0, 5)
	descriptions = append(descriptions, constructAttrValDescript("^$", AttributeValueNULL))
	descriptions = append(descriptions, constructAttrValDescript(`^-?\d+$`, AttributeValueINT))
	descriptions = append(descriptions, constructAttrValDescript(`^-?\d+\.\d+$`, AttributeValueDOUBLE))
	descriptions = append(descriptions, constructAttrValDescript(`^(true|false)$`, AttributeValueBOOL))
	descriptions = append(descriptions, constructAttrValDescript(`^\{"\w+":.+\}$`, AttributeValueMAP))
	descriptions = append(descriptions, constructAttrValDescript(`^\[.*\]$`, AttributeValueARRAY))
	return descriptions
}

func constructAttrValDescript(regex string, attrType AttributeValueType) *attrValDescript {
	regexc := regexp.MustCompile(regex)
	return &attrValDescript{
		regex:    regexc,
		attrType: attrType,
	}
}

func DetermineValueType(value string, omitSimpleTypes bool) AttributeValueType {
	if omitSimpleTypes {
		for _, desc := range complexAttrValDescriptions {
			if desc.regex.MatchString(value) {
				return desc.attrType
			}
		}
	} else {
		for _, desc := range attrValDescriptions {
			if desc.regex.MatchString(value) {
				return desc.attrType
			}
		}
	}
	return AttributeValueSTRING
}

type AttributeValueType int

const (
	AttributeValueNULL AttributeValueType = iota
	AttributeValueSTRING
	AttributeValueINT
	AttributeValueDOUBLE
	AttributeValueBOOL
	AttributeValueMAP
	AttributeValueARRAY
)

func (avt AttributeValueType) String() string {
	switch avt {
	case AttributeValueNULL:
		return "NULL"
	case AttributeValueSTRING:
		return "STRING"
	case AttributeValueBOOL:
		return "BOOL"
	case AttributeValueINT:
		return "INT"
	case AttributeValueDOUBLE:
		return "DOUBLE"
	case AttributeValueMAP:
		return "MAP"
	case AttributeValueARRAY:
		return "ARRAY"
	}
	return ""
}

func UnmarshalTraceJSON(data []byte) (result []byte, e error) {
	result = make([]byte, 16)
	e = unmarshalJSON(result, data)
	return result, e
}

func UnmarshalSpanJSON(data []byte) (result []byte, e error) {
	result = make([]byte, 8)
	e = unmarshalJSON(result, data)
	return result, e
}
