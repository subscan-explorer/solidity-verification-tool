package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"verify-golang/util"
)

const (
	mismatch = "mismatch"
	perfect  = "perfect"
	partial  = "partial"
)

var (
	ErrBytecodeNotFound       = errors.New("address not a contract or bytecode not found")
	InvalidValidInputMetadata = errors.New("invalid metadata")
	InvalidValidAddress       = errors.New("invalid address")
)

type VerificationRequest struct {
	Address         string `json:"address"`
	Metadata        string `json:"metadata"`
	Chain           int64  `json:"chain"`
	CompilerVersion string `json:"compilerVersion"`
}

type VerificationResponse struct {
	VerifiedStatus         string        `json:"verified_status"`
	Message                string        `json:"message"`
	Abi                    []interface{} `json:"abi,omitempty"`
	CreationBytecodeLength int           `json:"creation_bytecode_length"`
	ReviveVersion          string        `json:"revive_version,omitempty"`
}

// https://ardislu.dev/solc-standard-json-input-from-metadata
func verificationHandler(w http.ResponseWriter, r *http.Request) {
	var req VerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Metadata == "" || req.Address == "" || req.CompilerVersion == "" || req.Chain < 0 {
		respondError(w, InvalidValidInputMetadata)
		return
	}
	if !util.VerifyEthereumAddress(req.Address) {
		respondError(w, InvalidValidAddress)
		return
	}

	if !strings.HasPrefix(req.CompilerVersion, "v") {
		req.CompilerVersion = "v" + req.CompilerVersion
	}

	ctx := r.Context()
	if err := SolcManagerInstance.EnsureVersion(req.CompilerVersion); err != nil {
		respondError(w, err)
		return
	}

	// get bytecode from chain
	chainBytecode, err := req.fetchChainBytecode(ctx)
	if err != nil {
		respondError(w, err)
		return
	}
	if chainBytecode == "" {
		respondError(w, ErrBytecodeNotFound)
		return
	}

	inputJson, err := req.VerifyMetadata()
	if err != nil {
		respondError(w, err)
		return
	}
	util.Logger().Info(fmt.Sprintf("start compile contract %s with version %s", req.Address, req.CompilerVersion))
	compiledOutput, err := inputJson.recompileContract(ctx, req.CompilerVersion)
	if err != nil {
		util.Logger().Error(fmt.Errorf("compile contract %s with version %s failed: %s", req.Address, req.CompilerVersion, err.Error()))
		respondError(w, err)
		return
	}

	verified, err := req.compareBytecodes(ctx, chainBytecode, compiledOutput)

	if err != nil {
		respondError(w, err)
		return
	}

	if verified.Status == mismatch {
		respondError(w, fmt.Errorf("bytecode mismatch"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(VerificationResponse{VerifiedStatus: verified.Status,
		Message:                "ok",
		Abi:                    compiledOutput.Contracts[compiledOutput.CompileTarget][compiledOutput.ContractName].Abi,
		CreationBytecodeLength: len(compiledOutput.Contracts[compiledOutput.CompileTarget][compiledOutput.ContractName].Evm.Bytecode.Object),
		ReviveVersion:          compiledOutput.ReviveVersion,
	})
}

func respondError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(VerificationResponse{VerifiedStatus: mismatch, Message: err.Error()})
}
