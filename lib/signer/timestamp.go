package signer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/digitorus/timestamp"
)

func (context *SignContext) GetTSA(sign_content []byte) (timestamp_response []byte, err error) {
	sign_reader := bytes.NewReader(sign_content)
	ts_request, err := timestamp.CreateRequest(sign_reader, &timestamp.RequestOptions{
		Hash:         context.SignData.DigestAlgorithm,
		Certificates: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	ts_request_reader := bytes.NewReader(ts_request)
	req, err := http.NewRequest("POST", context.SignData.TSA.URL, ts_request_reader)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request (%s): %w", context.SignData.TSA.URL, err)
	}

	req.Header.Add("Content-Type", "application/timestamp-query")
	req.Header.Add("Content-Transfer-Encoding", "binary")

	if context.SignData.TSA.Username != "" && context.SignData.TSA.Password != "" {
		req.SetBasicAuth(context.SignData.TSA.Username, context.SignData.TSA.Password)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	code := 0

	if resp != nil {
		code = resp.StatusCode
	}

	if err != nil || (code < 200 || code > 299) {
		if err == nil {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read non success response: %w", err)
			}
			return nil, errors.New("non success response (" + strconv.Itoa(code) + "): " + string(body))
		}

		return nil, errors.New("non success response (" + strconv.Itoa(code) + ")")
	}

	defer resp.Body.Close()
	timestamp_response_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return timestamp_response_body, nil
}
