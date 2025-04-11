package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
)

type IMetadata interface {
	recompileContract(_ context.Context, version string) (*SolcOutput, error)
}

type SolcMetadata struct {
	Language string              `json:"language"`
	Sources  SourcesCode         `json:"sources"`
	Settings SolcMetadataSetting `json:"settings"`
	Output   *SolcMetadataOutput `json:"output,omitempty"`
	Compiler *map[string]string  `json:"compiler,omitempty"`
	Version  *float64            `json:"version,omitempty"`
}

// Format removes the compiler and version fields from the metadata
func (s *SolcMetadata) format() {
	s.Output = nil
	s.Compiler = nil
	s.Version = nil
	s.Settings.CompilationTarget = nil
	s.Settings.OutputSelection = map[string]map[string]interface{}{"*": {"*": []string{"abi", "evm.bytecode", "evm.deployedBytecode"}}}
}

func (s *SolcMetadata) PickComplicationTarget() (string, string) {
	for k, v := range s.Settings.CompilationTarget {
		return k, v
	}
	return "", ""
}

func (s *SolcMetadata) String() string {
	s.format()
	b, _ := json.Marshal(s)
	return string(b)
}

type SourcesCode map[string]SolcSources

type SolcMetadataSetting struct {
	Remappings []string `json:"remappings,omitempty"`
	Optimizer  struct {
		Enabled bool `json:"enabled"`
		Runs    int  `json:"runs"`
	} `json:"optimizer"`
	EvmVersion        string                 `json:"evmVersion"`
	Libraries         map[string]interface{} `json:"libraries,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CompilationTarget map[string]string      `json:"compilationTarget,omitempty"`
	OutputSelection   interface{}            `json:"outputSelection,omitempty"`
}

type SolcMetadataOutput struct {
	Abi                    []interface{} `json:"abi"`
	Devdoc                 interface{}   `json:"devdoc"`
	Userdoc                interface{}   `json:"userdoc"`
	CreationBytecodeLength int           `json:"creationBytecodeLength"` // custom, sourcify return
}

type SolcSources struct {
	Keccak256 string `json:"keccak256"`
	Content   string `json:"content"`
}

type SolcOutput struct {
	Contracts map[string]map[string]SolcContract `json:"contracts"`
	Errors    []struct {
		Component        string `json:"component"`
		ErrorCode        string `json:"errorCode"`
		FormattedMessage string `json:"formattedMessage"`
		Message          string `json:"message"`
		Severity         string `json:"severity"`
		SourceLocation   struct {
			End   int    `json:"end"`
			File  string `json:"file"`
			Start int    `json:"start"`
		} `json:"sourceLocation"`
		Type string `json:"type"`
	} `json:"errors"`
	ContractName  string
	CompileTarget string
}

func (o *SolcOutput) PickDeployedBytesCode(compileTarget, contractName string) string {
	if compileTarget == "" {
		for target := range o.Contracts {
			for name := range o.Contracts[target] {
				return o.Contracts[target][name].Evm.DeployedBytecode.Object
			}
		}
	}
	return o.Contracts[compileTarget][contractName].Evm.DeployedBytecode.Object
}

func (o *SolcOutput) PickBytesCode(compileTarget, contractName string) string {
	if compileTarget == "" {
		for target := range o.Contracts {
			for name := range o.Contracts[target] {
				return o.Contracts[target][name].Evm.Bytecode.Object
			}
		}
	}
	return o.Contracts[compileTarget][contractName].Evm.Bytecode.Object
}

type SolcContract struct {
	Abi []any `json:"abi"`
	Evm struct {
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
		DeployedBytecode struct {
			Object string `json:"object"`
		}
	} `json:"evm"`
	Metadata any `json:"metadata,omitempty"`
}

func (s *SolcMetadata) recompileContract(_ context.Context, version string) (*SolcOutput, error) {
	solcPath := filepath.Join(SolcManagerInstance.cacheDir, version)

	cmd := exec.Command(solcPath, "--standard-json")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	var result SolcOutput
	result.CompileTarget, result.ContractName = s.PickComplicationTarget()
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create pipe fail: %v", err)
	}

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("start cmd fail %v", err)
	}

	if _, err = io.WriteString(stdinPipe, s.String()); err != nil {
		return nil, err
	}
	stdinPipe.Close()

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(stdoutBuf.Bytes(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func BytecodeWithoutMetadata(code string) string {
	if len(code) < 6 {
		return code
	}
	numericResult, err := strconv.ParseInt(code[len(code)-4:], 16, 64)
	if err != nil {
		return code
	}
	metadataSize := int((numericResult * 2) + 4)
	if metadataSize > len(code) {
		return code
	}
	return code[:len(code)-metadataSize]
}
