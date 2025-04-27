package util

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestDebug(t *testing.T) {
	var buf bytes.Buffer
	SetLogger(&buf)

	tests := []struct {
		input    interface{}
		expected string
	}{
		{"test string", "test string"},
		{[]byte("test bytes"), "test bytes"},
		{fmt.Errorf("test error"), "test error"},
		{struct{ Name string }{"test"}, `{
  "Name": "test"
}`},
	}

	for _, tt := range tests {
		Debug(tt.input)
		if !strings.Contains(buf.String(), tt.expected) {
			t.Errorf("expected: %s, got: %s", tt.expected, buf.String())
		}
		buf.Reset()
	}
}
