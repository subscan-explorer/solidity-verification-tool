package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"verify-golang/util"
)

type EthRpcRes struct {
	Result string
}

func (v *VerificationRequest) fetchChainBytecode(ctx context.Context) (string, error) {
	chain, ok := chainGroup[v.Chain]
	if !ok {
		return "", fmt.Errorf("network %d not supported", v.Chain)
	}
	randomId := rand.Intn(100000)
	data, err := util.PostWithJson(ctx, []byte(fmt.Sprintf(`{"id": %d,"jsonrpc": "2.0","params": ["%s","latest"],"method": "eth_getCode"}`, randomId, v.Address)), chain.Rpc[0])

	if err != nil {
		return "", err
	}
	var result EthRpcRes
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	return result.Result, nil
}

func (v *VerificationRequest) VerifyMetadata() (IMetadata, error) {
	var metadata SolcMetadata
	err := json.Unmarshal([]byte(v.Metadata), &metadata)
	if err != nil {
		return nil, InvalidValidInputMetadata
	}
	if len(metadata.Sources) == 0 {
		return nil, InvalidValidInputMetadata
	}
	for _, source := range metadata.Sources {
		if source.Content == "" {
			return nil, InvalidValidInputMetadata
		}
	}
	// detect if is revive metadata
	if metadata.ResolcVersion != "" {
		return &ReviveMetadata{metadata}, nil
	}
	return &metadata, nil
}

type SubscanRes struct {
	Code int `json:"code"`
	Data struct {
		CreationCode string `json:"creation_code"`
	}
	Message string `json:"message"`
}

func fetchCreateBytecode(ctx context.Context, address string, networkID int64) (string, error) {
	chain, ok := chainGroup[networkID]
	if !ok {
		return "", fmt.Errorf("network %d not supported", networkID)
	}
	data, err := util.PostWithJson(ctx, []byte(fmt.Sprintf(`{"address":"%s"}`, address)), chain.ContractFetchAddress)

	if err != nil {
		util.Logger().Error(fmt.Errorf("fetch create bytecode failed for address %s on network %d: %v", address, networkID, err))
		return "", err
	}
	var result SubscanRes
	err = json.Unmarshal(data, &result)
	if err != nil {
		util.Logger().Error(fmt.Errorf("unmarshal create bytecode response failed for address %s on network %d: %v", address, networkID, err))
		return "", err
	}
	if result.Code != 0 {
		return "", fmt.Errorf("fetch create bytecode failed: %s", result.Message)
	}

	return result.Data.CreationCode, nil
}

type Match struct {
	Status          string
	ConstructorArgs string
}

func (v *VerificationRequest) compareBytecodes(ctx context.Context, chainBytecode string, compiledOutput *SolcOutput) (*Match, error) {
	recompileDeployCodeWithLibraries := addLibraryAddresses(compiledOutput.PickDeployedBytesCode(compiledOutput.CompileTarget, compiledOutput.ContractName), chainBytecode).Replaced
	if util.TrimHex(recompileDeployCodeWithLibraries) == util.TrimHex(chainBytecode) {
		return &Match{Status: perfect}, nil
	}

	trimmedChainBytecode := util.TrimHex(BytecodeWithoutMetadata(chainBytecode))
	trimmedWithLibraries := util.TrimHex(BytecodeWithoutMetadata(recompileDeployCodeWithLibraries))
	if trimmedChainBytecode == util.TrimHex(trimmedWithLibraries) {
		return &Match{Status: partial}, nil
	}

	if len(trimmedChainBytecode) == len(trimmedWithLibraries) {
		createData, err := fetchCreateBytecode(ctx, v.Address, v.Chain)
		if err != nil {
			return &Match{Status: mismatch}, fmt.Errorf("fetch create bytecode failed: please retry later")
		}

		createData = util.TrimHex(createData)
		if len(createData) > 0 {
			recompileBytesCodeWithLibraries := addLibraryAddresses(compiledOutput.PickBytesCode(compiledOutput.CompileTarget, compiledOutput.ContractName), createData).Replaced
			encodedConstructorArgs := extractEncodedConstructorArgs(createData, recompileBytesCodeWithLibraries)
			if strings.HasPrefix(createData, BytecodeWithoutMetadata(recompileBytesCodeWithLibraries)) {
				return &Match{Status: perfect, ConstructorArgs: encodedConstructorArgs}, nil
			}
		}

	}
	return &Match{Status: mismatch}, nil
}

type addLibraryAddressesResult struct {
	Replaced   string
	LibraryMap map[string]string
}

func addLibraryAddresses(template, real string) addLibraryAddressesResult {
	const placeholderStart = "__$"
	const placeholderLength = 40
	libraryMap := make(map[string]string)
	replaced := template

	for {
		index := strings.Index(replaced, placeholderStart)
		if index == -1 {
			break
		}
		// Check if there's enough length left for a full placeholder
		if index+placeholderLength > len(replaced) {
			break
		}
		placeholder := replaced[index : index+placeholderLength]

		// Ensure real has enough length at the current index
		if index+placeholderLength > len(real) {
			panic("real string length insufficient for placeholder")
		}
		address := real[index : index+placeholderLength]

		// Store mapping
		libraryMap[placeholder] = address

		// Escape $ signs for regex
		regexStr := strings.Replace(placeholder, "__$", "__\\$", -1)
		regexStr = strings.Replace(regexStr, "$__", "\\$__", -1)

		// Replace all occurrences
		re := regexp.MustCompile(regexStr)
		replaced = re.ReplaceAllString(replaced, address)
	}

	return addLibraryAddressesResult{Replaced: replaced, LibraryMap: libraryMap}
}

func extractEncodedConstructorArgs(creationData string, compiledCreationBytecode string) string {
	startIndex := strings.Index(creationData, compiledCreationBytecode)
	if len(creationData) <= startIndex+len(compiledCreationBytecode) {
		return ""
	}
	return "0x" + creationData[startIndex+len(compiledCreationBytecode):]
}
