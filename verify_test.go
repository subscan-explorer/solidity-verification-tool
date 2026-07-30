package main

import (
	"context"
	"encoding/json"
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
					Evm: SolcEVMOutput{
						Bytecode:         SolcBytecode{Object: "6000"},
						DeployedBytecode: SolcBytecode{Object: "6000"},
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

func TestCompareBytecodesSkipsLibraryPlaceholderBeyondChainBytecode(t *testing.T) {
	req := &VerificationRequest{}
	compiled := &SolcOutput{
		Contracts: map[string]map[string]SolcContract{
			"Other.sol": {
				"Other": {
					Evm: SolcEVMOutput{
						DeployedBytecode: SolcBytecode{
							Object: "6000__$1234567890123456789012345678901234$__",
						},
					},
				},
			},
			"Target.sol": {
				"Target": {
					Evm: SolcEVMOutput{
						DeployedBytecode: SolcBytecode{Object: "6000"},
					},
				},
			},
		},
	}

	match, err := req.compareBytecodes(context.Background(), "0x6000", compiled)
	if err != nil {
		t.Fatalf("compareBytecodes returned error: %v", err)
	}
	if match == nil || match.Status != perfect {
		t.Fatalf("expected perfect match, got %#v", match)
	}
	if compiled.CompileTarget != "Target.sol" || compiled.ContractName != "Target" {
		t.Fatalf("expected Target.sol:Target, got %s:%s", compiled.CompileTarget, compiled.ContractName)
	}
}

func TestCompareBytecodesMasksImmutableReferencesAndIgnoresMetadata(t *testing.T) {
	req := &VerificationRequest{}
	compiled := &SolcOutput{
		Contracts: map[string]map[string]SolcContract{
			"Executor.sol": {
				"Executor": {
					Evm: SolcEVMOutput{
						Bytecode: SolcBytecode{Object: "6000"},
						DeployedBytecode: SolcBytecode{
							Object: "6000600056ccdd0002",
							ImmutableReferences: ImmutableReferences{
								"3": {
									{Start: 1, Length: 1},
									{Start: 3, Length: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	match, err := req.compareBytecodes(context.Background(), "0x6049604956aabb0002", compiled)
	if err != nil {
		t.Fatalf("compareBytecodes returned error: %v", err)
	}
	if match.Status != partial {
		t.Fatalf("expected partial match, got %q", match.Status)
	}
	if compiled.CompileTarget != "Executor.sol" || compiled.ContractName != "Executor" {
		t.Fatalf("expected compile target Executor.sol:Executor, got %s:%s", compiled.CompileTarget, compiled.ContractName)
	}
}

func TestCompareBytecodesAllowsConstructorArgsWhenCreationCodeUnavailable(t *testing.T) {
	const constructorArgs = "0000000000000000000000000000000000000000000000000000000000000049"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"creation_code":""},"message":"Success"}`))
	}))
	defer server.Close()

	req := &VerificationRequest{
		Address:         "0xd633d8d1ceee8c8252196d44857c0f41b8dcb0d9",
		Chain:           222222,
		CompilerVersion: "v0.8.28",
		ConstructorArgs: constructorArgs,
	}
	oldChainGroup := chainGroup
	chainGroup = map[int64]ChainInfo{
		222222: {
			ContractFetchAddress: server.URL,
		},
	}
	t.Cleanup(func() {
		chainGroup = oldChainGroup
	})

	compiled := &SolcOutput{
		Contracts: map[string]map[string]SolcContract{
			"Executor.sol": {
				"Executor": {
					Evm: SolcEVMOutput{
						Bytecode: SolcBytecode{Object: "6000"},
						DeployedBytecode: SolcBytecode{
							Object: "6000600056ccdd0002",
							ImmutableReferences: ImmutableReferences{
								"3": {
									{Start: 1, Length: 1},
									{Start: 3, Length: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	match, err := req.compareBytecodes(context.Background(), "0x6049604956aabb0002", compiled)
	if err != nil {
		t.Fatalf("compareBytecodes returned error: %v", err)
	}
	if match.Status != partial {
		t.Fatalf("expected partial match, got %q", match.Status)
	}
	if match.ConstructorArgs != "0x"+constructorArgs {
		t.Fatalf("expected constructor args %q, got %q", "0x"+constructorArgs, match.ConstructorArgs)
	}
}

func TestMatchDeployedBytecodeMasksImmutableReferencesWithMatchingMetadata(t *testing.T) {
	status := matchDeployedBytecode("0x6049604956aabb0002", "6000600056aabb0002", ImmutableReferences{
		"3": {
			{Start: 1, Length: 1},
			{Start: 3, Length: 1},
		},
	})
	if status != perfect {
		t.Fatalf("expected perfect match, got %q", status)
	}
}

func TestMatchDeployedBytecodeRejectsNonImmutableDifference(t *testing.T) {
	status := matchDeployedBytecode("0x6049614956aabb0002", "6000600056ccdd0002", ImmutableReferences{
		"3": {
			{Start: 1, Length: 1},
			{Start: 3, Length: 1},
		},
	})
	if status != "" {
		t.Fatalf("expected mismatch, got %q", status)
	}
}

func TestSolcOutputParsesImmutableReferences(t *testing.T) {
	var output SolcOutput
	err := json.Unmarshal([]byte(`{
		"contracts": {
			"Executor.sol": {
				"Executor": {
					"evm": {
						"deployedBytecode": {
							"object": "6000",
							"immutableReferences": {
								"3": [
									{"start": 454, "length": 2},
									{"start": 823, "length": 2}
								]
							}
						}
					}
				}
			}
		}
	}`), &output)
	if err != nil {
		t.Fatalf("unmarshal solc output: %v", err)
	}

	references := output.Contracts["Executor.sol"]["Executor"].Evm.DeployedBytecode.ImmutableReferences["3"]
	if len(references) != 2 {
		t.Fatalf("expected 2 immutable references, got %d", len(references))
	}
	if references[0].Start != 454 || references[0].Length != 2 || references[1].Start != 823 || references[1].Length != 2 {
		t.Fatalf("unexpected immutable references: %+v", references)
	}
}

func TestExtractEncodedConstructorArgsWithDifferentMetadataHash(t *testing.T) {
	creationData := "6000aabb00021234"
	compiledCreationBytecode := "6000ccdd0002"

	got := extractEncodedConstructorArgs(creationData, compiledCreationBytecode)
	if got != "0x1234" {
		t.Fatalf("expected constructor args 0x1234, got %q", got)
	}
}
