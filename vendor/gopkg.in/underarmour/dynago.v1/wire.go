package dynago

import (
	"encoding/base64"
	"strconv"
	"time"
)

func wireEncode(value interface{}) interface{} {
	// This is somewhat optimized based on what we expect are the most common types.
	switch v := value.(type) {
	case string:
		return &wireString{v}
	case int:
		return &wireNumber{strconv.Itoa(v)}
	case int64:
		return &wireNumber{strconv.FormatInt(v, 10)}
	case bool:
		return &wireBool{v}
	case float64:
		return &wireNumber{strconv.FormatFloat(v, 'g', -1, 64)}
	case Number:
		return &wireNumber{string(v)}
	case Document:
		output := make(map[string]interface{}, len(v))
		for key, val := range v {
			output[key] = wireEncode(val)
		}
		return &wireMap{output}
	case []byte:
		return &wireBinary{v}
	case StringSet:
		return &wireStringSet{v}
	case NumberSet:
		return &wireNumberSet{v}
	case BinarySet:
		return &wireBinarySet{v}
	case List:
		encList := make([]interface{}, len(v))
		for i, raw := range v {
			encList[i] = wireEncode(raw)
		}
		return &wireList{encList}
	case map[string]interface{}:
		return wireEncode(Document(v))
	case int32, int16, int8:
		return &wireNumber{strconv.FormatInt(anyInt(v), 10)}
	case uint, uint64, uint32, uint16, uint8:
		return &wireNumber{strconv.FormatUint(anyUint(v), 10)}
	case time.Time:
		return wireEncodeTime(v)
	case *time.Time:
		return wireEncodeTime(*v)
	default:
		panic(v)
	}
}

func wireEncodeTime(t time.Time) interface{} {
	if t.Location() != time.UTC {
		panic("Times must be provided as UTC")
	}
	return &wireString{t.Format(iso8601compact)}
}

type wireString struct {
	S string
}

type wireStringSet struct {
	SS []string
}

type wireNumber struct {
	N string
}

type wireNumberSet struct {
	NS []string
}

type wireBool struct {
	BOOL bool `json:"BOOL"`
}

// Bonus! we don't have to do anything at all for binary. base64 is done by encoding/json.
type wireBinary struct {
	B []byte
}

type wireBinarySet struct {
	BS [][]byte
}

type wireList struct {
	L []interface{}
}

type wireMap struct {
	M map[string]interface{}
}

func wireDecode(original interface{}) interface{} {
	vv, ok := original.(map[string]interface{})
	if !ok {
		panic(original) // XXX DEBUG TODO
	}
	for typeCode, val := range vv {
		// TODO decode all the various types coming back.
		switch typeCode {
		case "S", "BOOL":
			return val
		case "NULL":
			return nil
		case "N":
			return Number(val.(string))
		case "NS":
			return wireDecodeNumberSet(val)
		case "SS":
			return wireDecodeStringSet(val)
		case "L":
			return wireDecodeList(val)
		case "M":
			return wireDecodeMap(val)
		case "B":
			return wireDecodeBinary(val)
		case "BS":
			return wireDecodeBinarySet(val)
		}
	}
	return nil
}

func wireDecodeNumberSet(val interface{}) interface{} {
	valSlice := val.([]interface{})
	resultSlice := make(NumberSet, len(valSlice))
	for i, v := range valSlice {
		resultSlice[i] = v.(string)
	}
	return resultSlice
}

func wireDecodeStringSet(val interface{}) interface{} {
	valSlice := val.([]interface{})
	resultSlice := make(StringSet, len(valSlice))
	for i, v := range valSlice {
		resultSlice[i] = v.(string)
	}
	return resultSlice
}

func wireDecodeList(val interface{}) interface{} {
	valSlice := val.([]interface{})
	resultSlice := make(List, len(valSlice))
	for i, v := range valSlice {
		resultSlice[i] = wireDecode(v)
	}
	return resultSlice
}

func wireDecodeMap(val interface{}) interface{} {
	m := val.(map[string]interface{})
	output := make(Document, len(m))
	for key, val := range m {
		output[key] = wireDecode(val)
	}
	return output
}

func wireDecodeBinary(val interface{}) []byte {
	s := val.(string)
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return buf
}

func wireDecodeBinarySet(val interface{}) interface{} {
	valSlice := val.([]interface{})
	resultSlice := make(BinarySet, len(valSlice))
	for i, v := range valSlice {
		resultSlice[i] = wireDecodeBinary(v)
	}
	return resultSlice
}

func anyInt(input interface{}) int64 {
	switch v := input.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	}
	panic("Unknown int type")
}

func anyUint(input interface{}) uint64 {
	switch v := input.(type) {
	case uint:
		return uint64(v)
	case uint64:
		return v
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	}
	panic("unknown uint type")
}
