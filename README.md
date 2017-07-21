orderedmap
==========

Orderedmap works like a `map[string]interface{}` but preserves the order in which keys have been added.
It has been optimized performance-wise while still keeping memory usage at a minimum. All operations have *O(1)* time complexity.

Orderedmap is useful for marshalling to JSON and can be easy marshalled using golang's `encoding/json` package.

## Full example code
```golang
package main

import "fmt"
import "encoding/json"

import om "github.com/maurice2k/orderedmap"

func main() {
    status := om.NewOrderedMap(  // initializes a new odered map
        &om.KV{"http", om.NewOrderedMap(  // key-value with sub ordered map as value
            &om.KV{"code", 200},  // simple key-value
            &om.KV{"message", "OK"},
        )})

    response := om.NewOrderedMap().  // now with Setter
        Set("success", true).
        Set("status", status)

    contentResponse := om.NewOrderedMap(
        &om.KV{"user", om.NewOrderedMap(
            &om.KV{"id", 1337},
            &om.KV{"name", "John Doe"},
        )})

    contentResponse.Set("success", false)  // will be skipped in Append(..., false) call

    // append content to response without overwriting existing keys
    // (second parameter set to false)
    response.Append(contentResponse, false)

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
