package apiclients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
)

type APIError struct {
	StatusCode int
	Raw        string
	Render     *apitypes.RenderErrorResponse
}

func ErrorFromResponse(resp *http.Response, body []byte) *APIError {
	if resp == nil {
		return &APIError{Raw: strings.TrimSpace(string(body))}
	}

	return decodeError(resp.StatusCode, body)
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "signoz api: HTTP %d", e.StatusCode)

	if e.Render == nil {
		if e.Raw != "" {
			fmt.Fprintf(&b, ": %s", e.Raw)
		}
		return b.String()
	}

	body := e.Render.Error
	if body.Code != "" {
		fmt.Fprintf(&b, " (%s)", body.Code)
	}
	if body.Message != "" {
		fmt.Fprintf(&b, ": %s", body.Message)
	}
	if body.Errors != nil {
		for _, d := range *body.Errors {
			if d.Message != nil && *d.Message != "" {
				fmt.Fprintf(&b, " | %s", *d.Message)
			}
		}
	}

	return b.String()
}

func decodeError(status int, raw []byte) *APIError {
	out := &APIError{StatusCode: status, Raw: strings.TrimSpace(string(raw))}

	var re apitypes.RenderErrorResponse
	if err := json.Unmarshal(raw, &re); err != nil {
		return out
	}

	if re.Error.Code != "" || re.Error.Message != "" || (re.Error.Errors != nil && len(*re.Error.Errors) > 0) {
		out.Render = &re
	}

	return out
}
