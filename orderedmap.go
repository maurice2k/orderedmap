// Copyright 2017-2026 Moritz Fain
// Moritz Fain <moritz@mertinkat.net>

// Package orderedmap provides a JSON-friendly map that preserves insertion
// order while keeping key lookups constant time.
package orderedmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
)

// KV is a key-value entry in an OrderedMap.
type KV struct {
	Key   string
	Value any
}

// OrderedMap stores string keys in insertion order.
//
// The zero value is ready to use.
type OrderedMap struct {
	kvList    []*KV
	idxLookup map[string]int
}

// New creates an ordered map from the provided entries.
func New(kvList ...KV) *OrderedMap {
	om := &OrderedMap{
		idxLookup: make(map[string]int, len(kvList)),
	}

	for _, kv := range kvList {
		om.Set(kv.Key, kv.Value)
	}
	return om
}

// NewOrderedMap creates an ordered map from the provided entries.
//
// Deprecated: use New with value entries.
func NewOrderedMap(kvList ...*KV) *OrderedMap {
	om := &OrderedMap{
		idxLookup: make(map[string]int, len(kvList)),
	}

	for _, kv := range kvList {
		if kv != nil {
			om.Set(kv.Key, kv.Value)
		}
	}
	return om
}

// Set stores value for key. Existing keys keep their original position.
func (om *OrderedMap) Set(key string, value any) *OrderedMap {
	om.ensureInit()

	if idx, ok := om.idxLookup[key]; ok {
		om.kvList[idx].Value = value
		return om
	}

	om.idxLookup[key] = len(om.kvList)
	om.kvList = append(om.kvList, &KV{Key: key, Value: value})
	return om
}

// Get returns key's value, or nil if key does not exist.
//
// Use Lookup when nil is a valid stored value.
func (om *OrderedMap) Get(key string) any {
	value, _ := om.Lookup(key)
	return value
}

// Lookup returns key's value and whether the key exists.
func (om *OrderedMap) Lookup(key string) (any, bool) {
	if om == nil {
		return nil, false
	}

	idx, ok := om.idxLookup[key]
	if !ok || om.kvList[idx] == nil {
		return nil, false
	}
	return om.kvList[idx].Value, true
}

// Exists reports whether key exists.
func (om *OrderedMap) Exists(key string) (ok bool) {
	if om == nil {
		return false
	}

	_, ok = om.idxLookup[key]
	return
}

// Delete removes key from the map.
func (om *OrderedMap) Delete(key string) {
	if om == nil {
		return
	}

	if idx, ok := om.idxLookup[key]; ok {
		delete(om.idxLookup, key)
		om.kvList[idx] = nil
	}
}

// Keys returns keys in insertion order.
func (om *OrderedMap) Keys() []string {
	if om == nil {
		return nil
	}

	keys := make([]string, 0, om.Len())
	for _, kv := range om.kvList {
		if kv != nil {
			keys = append(keys, kv.Key)
		}
	}
	return keys
}

// GetKeys returns keys in insertion order.
//
// Deprecated: use Keys.
func (om *OrderedMap) GetKeys() []string {
	return om.Keys()
}

// Entries returns key-value entries in insertion order.
func (om *OrderedMap) Entries() []KV {
	if om == nil {
		return nil
	}

	kvList := make([]KV, 0, om.Len())
	for _, kv := range om.kvList {
		if kv != nil {
			kvList = append(kvList, *kv)
		}
	}
	return kvList
}

// GetList returns key-value entries in insertion order.
//
// Deprecated: use Entries.
func (om *OrderedMap) GetList() []KV {
	return om.Entries()
}

// Append appends entries from newOm to om.
//
// Existing keys are overwritten only when overwrite is true.
func (om *OrderedMap) Append(newOm *OrderedMap, overwrite bool) *OrderedMap {
	if newOm == nil {
		return om
	}

	for _, kv := range newOm.Entries() {
		if !overwrite && om.Exists(kv.Key) {
			continue
		}
		om.Set(kv.Key, kv.Value)
	}
	return om
}

// Len returns the number of entries.
func (om *OrderedMap) Len() int {
	if om == nil {
		return 0
	}
	return len(om.idxLookup)
}

// All returns an iterator over entries in insertion order.
func (om *OrderedMap) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		if om == nil {
			return
		}

		for _, kv := range om.kvList {
			if kv == nil {
				continue
			}
			if !yield(kv.Key, kv.Value) {
				return
			}
		}
	}
}

// String returns the compact JSON representation of om.
func (om *OrderedMap) String() string {
	data, _ := om.MarshalJSON()
	return string(data)
}

// MarshalJSON implements json.Marshaler.
func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	if om == nil {
		return []byte("null"), nil
	}

	buffer := bytes.NewBufferString("{")

	first := true
	for _, kv := range om.kvList {
		if kv == nil {
			continue
		}

		if !first {
			buffer.WriteString(",")
		}

		jsonKey, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		jsonValue, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buffer.Write(jsonKey)
		buffer.WriteByte(':')
		buffer.Write(jsonValue)

		first = false
	}

	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

// UnmarshalJSON implements json.Unmarshaler.
//
// Nested JSON objects are decoded as *OrderedMap values.
func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	value, err := decodeJSONValue(decoder)
	if err != nil {
		return err
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return err
	}

	if value == nil {
		*om = OrderedMap{}
		return nil
	}

	decoded, ok := value.(*OrderedMap)
	if !ok {
		return fmt.Errorf("orderedmap: expected JSON object or null")
	}

	*om = *decoded
	return nil
}

// MarshalJSON implements json.Marshaler.
func (kv KV) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	jsonKey, err := json.Marshal(kv.Key)
	if err != nil {
		return nil, err
	}
	jsonValue, err := json.Marshal(kv.Value)
	if err != nil {
		return nil, err
	}

	buffer.Write(jsonKey)
	buffer.WriteByte(':')
	buffer.Write(jsonValue)

	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

func (om *OrderedMap) ensureInit() {
	if om.idxLookup == nil {
		om.idxLookup = make(map[string]int)
	}
}

func decodeJSONValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	return decodeJSONToken(decoder, token)
}

func decodeJSONToken(decoder *json.Decoder, token json.Token) (any, error) {
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}

	switch delim {
	case '{':
		om := New()
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}

			key, ok := keyToken.(string)
			if !ok {
				return nil, fmt.Errorf("orderedmap: expected object key")
			}

			value, err := decodeJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			om.Set(key, value)
		}

		if err := consumeDelim(decoder, '}'); err != nil {
			return nil, err
		}
		return om, nil

	case '[':
		values := make([]any, 0)
		for decoder.More() {
			value, err := decodeJSONValue(decoder)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}

		if err := consumeDelim(decoder, ']'); err != nil {
			return nil, err
		}
		return values, nil

	default:
		return nil, fmt.Errorf("orderedmap: unexpected JSON delimiter %q", delim)
	}
}

func consumeDelim(decoder *json.Decoder, want json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	if token != want {
		return fmt.Errorf("orderedmap: expected JSON delimiter %q", want)
	}
	return nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	return fmt.Errorf("orderedmap: unexpected trailing JSON token %v", token)
}
