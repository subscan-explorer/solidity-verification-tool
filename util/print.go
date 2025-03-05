package util

import (
	"encoding/json"
	"fmt"
)

func Debug(i interface{}) {
	var val string
	switch i := i.(type) {
	case string:
		val = i
	case []byte:
		val = string(i)
	case error:
		val = i.Error()
	default:
		b, _ := json.MarshalIndent(i, "", "  ")
		val = string(b)
	}
	fmt.Println(val)
}
