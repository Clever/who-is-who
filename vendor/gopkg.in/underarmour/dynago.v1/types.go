package dynago

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

/*
A set of binary blobs.

Note that BinarySet doesn't guarantee ordering on retrieval.
*/
type BinarySet [][]byte

/*
Lists represent DynamoDB lists, which are functionally very similar to JSON
lists.  Like JSON lists, these lists are heterogeneous, which means that the
elements of the list can be any valid value type, which includes other lists,
documents, numbers, strings, etc.
*/
type List []interface{}

/*
Return a copy of this list with all elements coerced as Documents.

It's very common to use lists in dynago where all elements in the list are
a Document. For that reason, this method is provided as a convenience to
get back your list as a list of documents.

If any element in the List is not a document, this will error.
As a convenience, even when it errors, a slice containing any elements
preceding the one which errored as documents will be given.
*/
func (l List) AsDocumentList() ([]Document, error) {
	docs := make([]Document, len(l))
	for i, listItem := range l {
		if doc, ok := listItem.(Document); !ok {
			return docs[:i], fmt.Errorf("item at index %d was not a Document", i)
		} else {
			docs[i] = doc
		}
	}
	return docs, nil
}

/*
Represents a number.

DynamoDB returns numbers as a string representation because they have a single
high-precision number type that can take the place of integers, floats, and
decimals for the majority of types.

This method has helpers to get the value of this number as one of various
Golang numeric type.
*/
type Number string

func (n Number) IntVal() (int, error) {
	return strconv.Atoi(string(n))
}

func (n Number) Int64Val() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

func (n Number) Uint64Val() (uint64, error) {
	return strconv.ParseUint(string(n), 10, 64)
}

func (n Number) FloatVal() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

/*
A set of numbers.
*/
type NumberSet []string

// Represents an entire document structure composed of keys and dynamo value
type Document map[string]interface{}

func (d Document) MarshalJSON() ([]byte, error) {
	output := make(map[string]interface{}, len(d))
	for key, val := range d {
		if v := reflect.ValueOf(val); !isEmptyValue(v) {
			output[key] = wireEncode(val)
		}
	}
	return json.Marshal(output)
}

func (d *Document) UnmarshalJSON(buf []byte) error {
	raw := make(map[string]interface{})
	err := json.Unmarshal(buf, &raw)
	if err != nil {
		return err
	}
	if *d == nil {
		*d = make(Document)
	}
	dd := *d

	for key, val := range raw {
		dd[key] = wireDecode(val)
	}
	return nil
}

/*
Helper to get a key from document as a List.

If value at key is nil, returns a nil list.
If value at key is not a List, will panic.
*/
func (d Document) GetList(key string) List {
	if d[key] != nil {
		return d[key].(List)
	} else {
		return nil
	}
}

// Helper to get a string from a document.
func (d Document) GetString(key string) string {
	if d[key] != nil {
		return d[key].(string)
	} else {
		return ""
	}
}

// Helper to get a Number from a document.
func (d Document) GetNumber(key string) Number {
	if d[key] != nil {
		return d[key].(Number)
	} else {
		return Number("")
	}
}

/*
Helper to get a key from a document as a StringSet.

If value at key does not exist; returns an empty StringSet.
If it exists but is not a StringSet, panics.
*/
func (d Document) GetStringSet(key string) StringSet {
	if d[key] != nil {
		return d[key].(StringSet)
	} else {
		return StringSet{}
	}
}

/*
Helper to get a Time from a document.

If the value is omitted from the DB, or an empty string, then the return
is nil. If the value fails to parse as iso8601, then this method panics.
*/
func (d Document) GetTime(key string) (t *time.Time) {
	val := d[key]
	if val != nil {
		s := val.(string)
		parsed, err := time.ParseInLocation(iso8601compact, s, time.UTC)
		if err != nil {
			panic(err)
		}
		t = &parsed
	}
	return t
}

// Allow a document to be used to specify params
func (d Document) AsParams() (params []Param) {
	for key, val := range d {
		params = append(params, Param{key, val})
	}
	return
}

/*
Gets the value at the key as a boolean.

If the value does not exist in this Document, returns false.
If the value is the nil interface, also returns false.
If the value is a bool, returns the value of the bool.
If the value is a Number, returns true if value is non-zero.
For any other values, panics.
*/
func (d Document) GetBool(key string) bool {
	if v := d[key]; v != nil {
		switch val := v.(type) {
		case bool:
			return val
		case Number:
			if res, err := val.IntVal(); err != nil {
				panic(err)
			} else if res == 0 {
				return false
			} else {
				return true
			}
		default:
			panic(v)
		}
	} else {
		return false
	}
}

// Helper to build a hash key.
func HashKey(name string, value interface{}) Document {
	return Document{name: value}
}

// Helper to build a hash-range key.
func HashRangeKey(hashName string, hashVal interface{}, rangeName string, rangeVal interface{}) Document {
	return Document{
		hashName:  hashVal,
		rangeName: rangeVal,
	}
}

type Param struct {
	Key   string
	Value interface{}
}

// Allows a solo Param to also satisfy the Params interface
func (p Param) AsParams() []Param {
	return []Param{p}
}

/*
Anything which implements Params can be used as expression parameters for
dynamodb expressions.

DynamoDB item queries using expressions can be provided parameters in a number
of handy ways:
	.Param(":k1", v1).Param(":k2", v2)
	-or-
	.Params(Param{":k1", v1}, Param{":k2", v2})
	-or-
	.FilterExpression("...", Param{":k1", v1}, Param{":k2", v2})
	-or-
	.FilterExpression("...", Document{":k1": v1, ":k2": v2})
Or any combination of Param, Document, or potentially other custom types which
provide the Params interface.
*/
type Params interface {
	AsParams() []Param
}

/*
Store a set of strings.

Sets in DynamoDB do not guarantee any ordering, so storing and retrieving a
StringSet may not give you back the same order you put it in. The main
advantage of using sets in DynamoDB is using atomic updates with ADD and DELETE
in your UpdateExpression.
*/
type StringSet []string

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}
