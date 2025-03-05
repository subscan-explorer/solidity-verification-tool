// main.go
package main

import (
	"net/http"
	"verify-golang/util"
)

func init() {
	fetchChainInfo()
}

func main() {
	http.HandleFunc("/verify", verificationHandler)
	util.Logger().Info("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
