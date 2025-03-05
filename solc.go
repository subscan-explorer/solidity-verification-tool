package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"verify-golang/util"
)

type SolcManager struct {
	versions sync.Map
	cacheDir string
}

func NewSolcManager() *SolcManager {
	// current code dir
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	staticDir := filepath.Join(dir, "static")
	return &SolcManager{cacheDir: staticDir}
}

func (sm *SolcManager) EnsureVersion(version string) error {
	if _, ok := sm.versions.Load(version); ok {
		return nil
	}

	versionDir := filepath.Join(sm.cacheDir, version)
	if _, err := os.Stat(versionDir); err == nil {
		sm.versions.Store(version, versionDir)
		return nil
	}

	util.Logger().Info(fmt.Sprintf("Start Downloading solc bin %s", version))
	if err := sm.downloadSolc(version); err != nil {
		return err
	}
	sm.versions.Store(versionDir, versionDir)
	return nil
}

const (
	// github solc repo
	GithubSolcRepoLinux = "https://github.com/ethereum/solc-bin/raw/gh-pages/linux-amd64/solc-linux-amd64-"
	GithubSolcRepoMacos = "https://github.com/ethereum/solc-bin/raw/gh-pages/macosx-amd64/solc-macosx-amd64-"
)

func (sm *SolcManager) downloadSolc(version string) error {
	repo := GithubSolcRepoLinux
	if runtime.GOOS == "darwin" {
		repo = GithubSolcRepoMacos
	}
	// https://raw.githubusercontent.com/ethereum/solc-bin/refs/heads/gh-pages/macosx-amd64/solc-macosx-amd64-v0.3.6%2Bcommit.988fe5e5
	url := fmt.Sprintf("%s%s", repo, version)

	util.Logger().Info(fmt.Sprintf("Downloading solc from %s", url))
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	solcFile := filepath.Join(sm.cacheDir, version)
	out, err := os.Create(solcFile)
	if err != nil {
		return err
	}
	_ = out.Chmod(0755)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
