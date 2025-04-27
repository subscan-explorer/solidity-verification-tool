package util

import (
	"testing"
)

func TestVerifyEthereumAddress(t *testing.T) {
	tests := []struct {
		accountId string
		expected  bool
	}{
		{"0x32Be343B94f860124dC4fEe278FDCBD38C102D88", true},
		{"32Be343B94f860124dC4fEe278FDCBD38C102D88", true},
		{"0x32Be343B94f860124dC4fEe278FDCBD38C102D8", false},
		{"", false},
		{"null", false},
	}

	for _, tt := range tests {
		result := VerifyEthereumAddress(tt.accountId)
		if result != tt.expected {
			t.Errorf("VerifyEthereumAddress(%s) = %v; want %v", tt.accountId, result, tt.expected)
		}
	}
}

func TestAddHex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"32Be343B94f860124dC4fEe278FDCBD38C102D88", "0x32be343b94f860124dc4fee278fdcbd38c102d88"},
		{"0x32Be343B94f860124dC4fEe278FDCBD38C102D88", "0x32Be343B94f860124dC4fEe278FDCBD38C102D88"},
		{"", ""},
		{"null", ""},
	}

	for _, tt := range tests {
		result := AddHex(tt.input)
		if result != tt.expected {
			t.Errorf("AddHex(%s) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}

func TestTrimHex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0x32Be343B94f860124dC4fEe278FDCBD38C102D88", "32Be343B94f860124dC4fEe278FDCBD38C102D88"},
		{"32Be343B94f860124dC4fEe278FDCBD38C102D88", "32Be343B94f860124dC4fEe278FDCBD38C102D88"},
		{"", ""},
	}

	for _, tt := range tests {
		result := TrimHex(tt.input)
		if result != tt.expected {
			t.Errorf("TrimHex(%s) = %v; want %v", tt.input, result, tt.expected)
		}
	}
}
