// main.go
package main

import (
	"log"
	"net/http"
	"os"
	"verify-golang/util"
)

func init() {
	fetchChainInfo()
	SolcManagerInstance = NewSolcManager()
	staticDir := SolcManagerInstance.cacheDir
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, 0755); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	http.HandleFunc("/verify", verificationHandler)
	util.Logger().Info("Server started on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
