// main.go
package main

import (
	"log"
	"net/http"
	"verify-golang/util"
)

func init() {
	fetchChainInfo()
}

func main() {
	http.HandleFunc("/verify", verificationHandler)
	util.Logger().Info("Server started on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
