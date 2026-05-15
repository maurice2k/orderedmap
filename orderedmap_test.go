package orderedmap_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	orderedmap "github.com/maurice2k/orderedmap"
)

func TestSetGetLookupAndOrder(t *testing.T) {
	var om orderedmap.OrderedMap
	firstKey := "first"

	om.Set(firstKey, 1).
		Set("second", 2).
		Set(firstKey, 3)

	if got := om.Get(firstKey); got != 3 {
		t.Fatalf("Get(%q) = %v, want 3", firstKey, got)
	}
	if value, ok := om.Lookup("missing"); ok || value != nil {
		t.Fatalf("Lookup missing = (%v, %t), want (nil, false)", value, ok)
	}
	if value, ok := om.Lookup(firstKey); !ok || value != 3 {
		t.Fatalf("Lookup(%q) = (%v, %t), want (3, true)", firstKey, value, ok)
	}
	if got, want := om.Keys(), []string{"first", "second"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	if got := om.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
}

func TestLookupDistinguishesStoredNilFromMissingKey(t *testing.T) {
	om := orderedmap.New(orderedmap.KV{Key: "nil", Value: nil})

	if got := om.Get("nil"); got != nil {
		t.Fatalf("Get stored nil = %v, want nil", got)
	}
	if value, ok := om.Lookup("nil"); !ok || value != nil {
		t.Fatalf("Lookup stored nil = (%v, %t), want (nil, true)", value, ok)
	}
	if value, ok := om.Lookup("missing"); ok || value != nil {
		t.Fatalf("Lookup missing = (%v, %t), want (nil, false)", value, ok)
	}
}

func TestEntriesAreCopies(t *testing.T) {
	om := orderedmap.New(orderedmap.KV{Key: "a", Value: 1})

	entries := om.Entries()
	entries[0].Value = 99

	if got := om.Get("a"); got != 1 {
		t.Fatalf("mutating Entries result changed map value to %v, want 1", got)
	}
}

func TestDeleteAndReinsert(t *testing.T) {
	om := orderedmap.New(
		orderedmap.KV{Key: "a", Value: 1},
		orderedmap.KV{Key: "b", Value: 2},
		orderedmap.KV{Key: "c", Value: 3},
	)

	om.Delete("b")
	om.Delete("missing")
	om.Set("b", 4)

	if got, want := om.Keys(), []string{"a", "c", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	if got := om.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}
	if got := om.Get("b"); got != 4 {
		t.Fatalf("Get(%q) = %v, want 4", "b", got)
	}
}

func TestAppend(t *testing.T) {
	target := orderedmap.New(
		orderedmap.KV{Key: "a", Value: 1},
		orderedmap.KV{Key: "b", Value: 2},
	)
	source := orderedmap.New(
		orderedmap.KV{Key: "b", Value: 22},
		orderedmap.KV{Key: "c", Value: 3},
	)

	target.Append(source, false)
	assertEntries(t, target, []orderedmap.KV{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
		{Key: "c", Value: 3},
	})

	target.Append(source, true)
	assertEntries(t, target, []orderedmap.KV{
		{Key: "a", Value: 1},
		{Key: "b", Value: 22},
		{Key: "c", Value: 3},
	})
}

func TestAllIteratesInOrderAndStopsEarly(t *testing.T) {
	om := orderedmap.New(
		orderedmap.KV{Key: "a", Value: 1},
		orderedmap.KV{Key: "b", Value: 2},
		orderedmap.KV{Key: "c", Value: 3},
	)

	var keys []string
	for key := range om.All() {
		keys = append(keys, key)
		if len(keys) == 2 {
			break
		}
	}

	if got, want := keys, []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("iterated keys = %v, want %v", got, want)
	}
}

func TestMarshalJSONPreservesOrder(t *testing.T) {
	om := orderedmap.New(
		orderedmap.KV{Key: "b", Value: 2},
		orderedmap.KV{Key: "a", Value: "two"},
	)

	data, err := json.Marshal(om)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	if got, want := string(data), `{"b":2,"a":"two"}`; got != want {
		t.Fatalf("json.Marshal = %s, want %s", got, want)
	}
	if got, want := om.String(), `{"b":2,"a":"two"}`; got != want {
		t.Fatalf("String() = %s, want %s", got, want)
	}
}

func TestKVMarshalJSON(t *testing.T) {
	data, err := json.Marshal(orderedmap.KV{Key: "a", Value: 1})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	if got, want := string(data), `{"a":1}`; got != want {
		t.Fatalf("json.Marshal(KV) = %s, want %s", got, want)
	}
}

func TestUnmarshalJSONPreservesOrderAndNestedObjects(t *testing.T) {
	var om orderedmap.OrderedMap
	input := []byte(`{"b":2,"a":{"y":true,"x":[{"z":1}]}}`)

	if err := json.Unmarshal(input, &om); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if got, want := om.Keys(), []string{"b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	if value, ok := om.Lookup("b"); !ok || value.(json.Number).String() != "2" {
		t.Fatalf("Lookup(%q) = (%v, %t), want json.Number(2), true", "b", value, ok)
	}

	nested, ok := om.Get("a").(*orderedmap.OrderedMap)
	if !ok {
		t.Fatalf("nested object decoded as %T, want *orderedmap.OrderedMap", om.Get("a"))
	}
	if got, want := nested.Keys(), []string{"y", "x"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nested Keys() = %v, want %v", got, want)
	}

	values, ok := nested.Get("x").([]any)
	if !ok || len(values) != 1 {
		t.Fatalf("nested array decoded as %#v, want one-element []any", nested.Get("x"))
	}
	item, ok := values[0].(*orderedmap.OrderedMap)
	if !ok {
		t.Fatalf("nested array object decoded as %T, want *orderedmap.OrderedMap", values[0])
	}
	if got, want := item.Keys(), []string{"z"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("array object Keys() = %v, want %v", got, want)
	}

	data, err := json.Marshal(&om)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if got, want := string(data), string(input); got != want {
		t.Fatalf("round-trip JSON = %s, want %s", got, want)
	}
}

func TestUnmarshalJSONRejectsNonObjectAndTrailingTokens(t *testing.T) {
	tests := map[string]string{
		"array":    `[]`,
		"trailing": `{} {}`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			var om orderedmap.OrderedMap
			if err := json.Unmarshal([]byte(input), &om); err == nil {
				t.Fatalf("json.Unmarshal(%s) succeeded, want error", input)
			}
		})
	}
}

func TestNewOrderedMapCompatibility(t *testing.T) {
	om := orderedmap.NewOrderedMap(
		&orderedmap.KV{Key: "a", Value: 1},
		nil,
		&orderedmap.KV{Key: "b", Value: 2},
	)

	if got, want := om.Keys(), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
}

func FuzzUnmarshalJSON(f *testing.F) {
	seeds := []string{
		`{}`,
		`null`,
		`{"a":1,"b":true,"c":null}`,
		`{"nested":{"z":[{"a":1},{"b":2}]}}`,
		`{"escaped":"\"\\\n"}`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var om orderedmap.OrderedMap
		if err := json.Unmarshal([]byte(input), &om); err != nil {
			return
		}

		data, err := json.Marshal(&om)
		if err != nil {
			t.Fatalf("json.Marshal after successful unmarshal returned error: %v", err)
		}

		var roundTrip orderedmap.OrderedMap
		if err := json.Unmarshal(data, &roundTrip); err != nil {
			t.Fatalf("round-trip unmarshal failed for %s: %v", data, err)
		}
	})
}

func assertEntries(t *testing.T, om *orderedmap.OrderedMap, want []orderedmap.KV) {
	t.Helper()

	if got := om.Entries(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Entries() = %#v, want %#v", got, want)
	}
}

func Example() {
	status := orderedmap.New(orderedmap.KV{Key: "http", Value: orderedmap.New(
		orderedmap.KV{Key: "code", Value: 200},
		orderedmap.KV{Key: "message", Value: "OK"},
	)})

	response := orderedmap.New().
		Set("success", true).
		Set("status", status)

	content := orderedmap.New(orderedmap.KV{Key: "user", Value: orderedmap.New(
		orderedmap.KV{Key: "id", Value: 1337},
		orderedmap.KV{Key: "name", Value: "John Doe"},
	)})

	response.Append(content, false)

	jsonData, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(jsonData))
	// Output:
	// {
	//   "success": true,
	//   "status": {
	//     "http": {
	//       "code": 200,
	//       "message": "OK"
	//     }
	//   },
	//   "user": {
	//     "id": 1337,
	//     "name": "John Doe"
	//   }
	// }
}
