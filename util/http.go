package util

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

func PostWithJson(ctx context.Context, data []byte, endpoint string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close() // nolint: errcheck
	body, _ := io.ReadAll(resp.Body)
	return body, nil
}
