package main

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	chainGroup map[int64]ChainInfo
	regOnce    sync.Once
)

type ChainInfo struct {
	Rpc                  []string `json:"rpc"`
	ContractFetchAddress string   `json:"contractFetchAddress"`

	// Deprecated, use metadata input instead
	// Revive bool `json:"revive"`
}

// Read chains.json to get all supported chain information
// Write to a local variable, only write once
func fetchChainInfo() map[int64]ChainInfo {
	regOnce.Do(func() {
		file, err := os.Open("chains.json")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		var chains map[int64]ChainInfo
		if err := json.NewDecoder(file).Decode(&chains); err != nil {
			panic(err)
		}
		chainGroup = chains
	})
	return chainGroup
}
