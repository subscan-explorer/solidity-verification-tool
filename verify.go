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

func (v *VerificationRequest) VerifyMetadata() (*SolcMetadata, error) {
	util.Debug(v.Metadata)
	var metadata SolcMetadata
	err := json.Unmarshal([]byte(v.Metadata), &metadata)
	util.Debug(metadata)
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
		return "", err
	}
	var result SubscanRes
	err = json.Unmarshal(data, &result)
	if err != nil {
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

func (v *VerificationRequest) compareBytecodes(ctx context.Context, deployedBytecode string, compiledOutput *SolcOutput) (*Match, error) {
	// add library Address
	withLibraries, _ := addLibraryAddresses(compiledOutput.PickDeployedBytesCode("", ""), deployedBytecode)

	if util.TrimHex(withLibraries) == util.TrimHex(deployedBytecode) {
		return &Match{Status: perfect}, nil
	}

	trimmedChainBytecode := util.TrimHex(BytecodeWithoutMetadata(deployedBytecode))
	util.Logger().Info(fmt.Sprintf("trimmedChainBytecode: %s", trimmedChainBytecode))
	trimmedWithLibraries := util.TrimHex(BytecodeWithoutMetadata(withLibraries))
	util.Logger().Info(fmt.Sprintf("trimmedWithLibraries: %s", trimmedWithLibraries))
	if trimmedChainBytecode == util.TrimHex(trimmedWithLibraries) {
		return &Match{Status: partial}, nil
	}

	if len(trimmedChainBytecode) == len(trimmedWithLibraries) {
		createData, err := fetchCreateBytecode(ctx, v.Address, v.Chain)
		if err != nil {
			return &Match{Status: mismatch}, err
		}

		createData = util.TrimHex(createData)
		if len(createData) > 0 {
			withLibraries, _ = addLibraryAddresses(compiledOutput.PickDeployedBytesCode("", ""), createData)
			encodedConstructorArgs := extractEncodedConstructorArgs(createData, withLibraries)
			util.Logger().Info(fmt.Sprintf("createData: %s", createData))
			util.Logger().Info(fmt.Sprintf("withLibraries: %s", withLibraries))

			if strings.HasPrefix(createData, withLibraries) {
				return &Match{Status: perfect, ConstructorArgs: encodedConstructorArgs}, nil
			}
		}

	}
	return &Match{Status: mismatch}, nil
}

func addLibraryAddresses(template, real string) (string, map[string]string) {
	const PlaceholderStart = "__$"
	const PlaceholderLength = 40

	libraryMap := make(map[string]string)

	index := strings.Index(template, PlaceholderStart)
	for index != -1 {
		placeholder := template[index : index+PlaceholderLength]
		address := real[index : index+PlaceholderLength]
		libraryMap[placeholder] = address
		regexCompatiblePlaceholder := strings.ReplaceAll(strings.ReplaceAll(placeholder, "__$", "__\\$"), "$__", "\\$__")
		regex := regexp.MustCompile(regexCompatiblePlaceholder)
		template = regex.ReplaceAllString(template, address)
		index = strings.Index(template, PlaceholderStart)
	}

	return template, libraryMap
}

func extractEncodedConstructorArgs(creationData string, compiledCreationBytecode string) string {
	startIndex := strings.Index(creationData, compiledCreationBytecode)
	if len(creationData) <= startIndex+len(compiledCreationBytecode) {
		return ""
	}
	return "0x" + creationData[startIndex+len(compiledCreationBytecode):]
}
