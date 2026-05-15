# orderedmap

`orderedmap` is a small Go package for JSON-friendly maps that preserve insertion order.
It behaves like a `map[string]any` for lookup and update, while iteration and JSON output follow the order in which keys were first inserted.

It is useful when generating stable, human-readable JSON with Go's `encoding/json` package.

## Install

```sh
go get github.com/maurice2k/orderedmap
```

## Complexity

- `Set`, `Get`, `Lookup`, `Exists`, `Delete`, and `Len` are O(1).
- `Keys`, `Entries`, `All`, `Append`, and JSON marshal/unmarshal are O(n).
- Deleting an entry preserves the relative order of the remaining entries. Re-inserting the same key appends it at the end.

## Full example code

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/maurice2k/orderedmap"
)

func main() {
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
}
```

## Output
```json
{
  "success": true,
  "status": {
    "http": {
      "code": 200,
      "message": "OK"
    }
  },
  "user": {
    "id": 1337,
    "name": "John Doe"
  }
}
```

## API notes

```go
om := orderedmap.New()

om.Set("id", 123)
value := om.Get("id")
value, ok := om.Lookup("id")
keys := om.Keys()
entries := om.Entries()

for key, value := range om.All() {
    _ = key
    _ = value
}
```

`Get` returns `nil` for missing keys. Use `Lookup` when `nil` is a valid stored value.

`json.Unmarshal` decodes JSON objects as `*orderedmap.OrderedMap` values, including nested objects.
Numbers are decoded as `json.Number`.

The old `NewOrderedMap`, `GetKeys`, and `GetList` names still exist as compatibility wrappers.
