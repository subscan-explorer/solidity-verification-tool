package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
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

	ctx := r.Context()
	sm := NewSolcManager()
	if err := sm.EnsureVersion(req.CompilerVersion); err != nil {
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
	compiledOutput, err := recompileContract(ctx, inputJson, req.CompilerVersion)
	if err != nil {
		util.Logger().Error(fmt.Sprintf("compile contract %s with version %s failed: %s", req.Address, req.CompilerVersion, err.Error()))
		respondError(w, err)
		return
	}

	verified, err := req.compareBytecodes(ctx, chainBytecode, compiledOutput)

	if err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(VerificationResponse{
		Verified: verified.Status != mismatch,
		Message:  "Verification completed",
	})
}

func respondError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(VerificationResponse{
		Verified: false,
		Message:  err.Error(),
	})
}
