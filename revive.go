package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"verify-golang/util"
)

type ReviveMetadata struct {
	SolcMetadata
}

func (s *ReviveMetadata) recompileContract(_ context.Context, version string) (*SolcOutput, error) {
	//  ./resolc --solc ./v0.8.17+commit.8df45f5f  --standard-json<example_input.json
	solcPath := filepath.Join(SolcManagerInstance.cacheDir, "resolc")
	fmt.Println("solcPath:", solcPath)
	cmd := exec.Command(solcPath, "--solc", filepath.Join(SolcManagerInstance.cacheDir, version), "--standard-json")
	fmt.Println(cmd)
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

// Release GitHub Release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func download() {
	util.Logger().Info("Start downloading latest resolc binary")
	fileName := downloadLatestResolc()
	err := extractAndSetExec(fileName, "static", strings.Replace(fileName, ".tar.gz", "", 1), "resolc")
	if err != nil {
		log.Fatal(err)
	}
}

func downloadLatestResolc() string {
	const repo = "paritytech/revive"
	fileName := "resolc-x86_64-unknown-linux-musl.tar.gz"
	if runtime.GOOS == "darwin" {
		fileName = "resolc-universal-apple-darwin.tar.gz"
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(apiURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("download failed：%d", resp.StatusCode))
	}

	var release Release
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		panic(err)
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == fileName {
			downloadURL = asset.DownloadURL
			break
		}
	}

	if downloadURL == "" {
		panic("file not found")
	}

	fileResp, err := http.Get(downloadURL)
	if err != nil {
		panic(err)
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("file download fail：%d", fileResp.StatusCode))
	}

	outFile, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, fileResp.Body)
	if err != nil {
		panic(err)
	}
	util.Logger().Info("resolc download success")
	return fileName
}

// extractAndSetExec uncompresses the tar.gz file and sets the executable permission for the specified file
func extractAndSetExec(src, dest, execFile, rename string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Name != execFile {
			continue
		}
		target := filepath.Join(dest, rename)
		switch header.Typeflag {
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()

			if filepath.Base(target) == execFile {
				if err := os.Chmod(target, 0755); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
