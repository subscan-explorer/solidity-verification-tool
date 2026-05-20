package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_fetchCreateBytecodeAddsSubscanAPIKeyHeader(t *testing.T) {
	const (
		apiKey   = "test-subscan-api-key"
		address  = "0x1234"
		network  = int64(999999)
		expected = "0x60006000"
	)

	t.Setenv("SUBSCAN_API_KEY", apiKey)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != apiKey {
			t.Fatalf("expected X-API-Key header %q, got %q", apiKey, got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		if got := string(body); got != `{"address":"0x1234"}` {
			t.Fatalf("expected request body %q, got %q", `{"address":"0x1234"}`, got)
		}

		_, _ = w.Write([]byte(`{"code":0,"data":{"creation_code":"` + expected + `"},"message":""}`))
	}))
	defer server.Close()

	oldChainGroup := chainGroup
	chainGroup = map[int64]ChainInfo{
		network: {
			ContractFetchAddress: server.URL,
		},
	}
	t.Cleanup(func() {
		chainGroup = oldChainGroup
	})

	got, err := fetchCreateBytecode(context.Background(), address, network)
	if err != nil {
		t.Fatalf("fetchCreateBytecode returned error: %v", err)
	}
	if got != expected {
		t.Fatalf("expected creation code %q, got %q", expected, got)
	}
}

func TestCompareBytecodesUsesProvidedConstructorArgs(t *testing.T) {
	const constructorArgs = "1234abcd"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"creation_code":"0x60001234abcd"},"message":""}`))
	}))
	defer server.Close()

	req := &VerificationRequest{
		Address:         "0x1234",
		Chain:           1,
		CompilerVersion: "v0.8.0",
		ConstructorArgs: constructorArgs,
	}
	oldChainGroup := chainGroup
	chainGroup = map[int64]ChainInfo{
		1: {
			ContractFetchAddress: server.URL,
		},
	}
	t.Cleanup(func() {
		chainGroup = oldChainGroup
	})
	compiled := &SolcOutput{
		Contracts: map[string]map[string]SolcContract{
			"Token.sol": {
				"Token": {
					Evm: struct {
						Bytecode struct {
							Object string "json:\"object\""
						} "json:\"bytecode\""
						DeployedBytecode struct {
							Object string "json:\"object\""
						}
					}{
						Bytecode: struct {
							Object string "json:\"object\""
						}{Object: "6000"},
						DeployedBytecode: struct {
							Object string "json:\"object\""
						}{Object: "6000"},
					},
				},
			},
		},
	}

	match, err := req.compareBytecodes(context.Background(), "0x6000", compiled)
	if err != nil {
		t.Fatalf("compareBytecodes returned error: %v", err)
	}
	if match == nil {
		t.Fatal("expected non-nil match")
	}
	if match.ConstructorArgs != "0x"+strings.ToLower(constructorArgs) {
		t.Fatalf("expected constructor args %q, got %q", "0x"+strings.ToLower(constructorArgs), match.ConstructorArgs)
	}
}
