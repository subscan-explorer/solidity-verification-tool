package util

import (
	"regexp"
	"strings"
)

func VerifyEthereumAddress(accountId string) bool {
	ethAddressRegex := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	return ethAddressRegex.MatchString(AddHex(accountId))
}

func AddHex(s string) string {
	if strings.TrimSpace(s) == "" || strings.TrimSpace(s) == "null" {
		return ""
	}
	if strings.HasPrefix(s, "0x") {
		return s
	}
	return strings.ToLower("0x" + s)
}

func TrimHex(s string) string {
	return strings.TrimPrefix(s, "0x")
}
