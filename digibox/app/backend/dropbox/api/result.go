package api

import (
	"encoding/json"
	"net/http"
)

type StatusType string

const (
	StatusSuccess StatusType = "success"
	StatusError   StatusType = "error"
)

type DropboxAPIResult struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Response   *http.Response
	Status     StatusType
	ErrorMsg   string
}

func ParseDropboxAPIError(body []byte) string {
	dropboxErrMsg := string(body)
	var dropboxErrObj map[string]interface{}

	if err := json.Unmarshal(body, &dropboxErrObj); err == nil {
		if msg, ok := dropboxErrObj["error_summary"]; ok {
			dropboxErrMsg = msg.(string)
		} else if msg, ok := dropboxErrObj["error_description"]; ok {
			dropboxErrMsg = msg.(string)
		}
	}

	return dropboxErrMsg
}

func buildDropboxAPIResult(resp *http.Response, body []byte) *DropboxAPIResult {
	result := &DropboxAPIResult{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		Response:   resp,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = StatusSuccess
		return result
	}

	result.Status = StatusError
	result.ErrorMsg = ParseDropboxAPIError(body)

	return result
}
