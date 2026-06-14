package apiclients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SigNoz/terraform-provider-signoz/internal/apitypes"
)

// maxRawBody caps APIError.Raw. The Raw path only fires when the response
// wasn't the JSON error envelope — typically a proxy's HTML page when upstream
// is down — so a few hundred bytes is enough to identify the failure without
// dumping a whole page into a Terraform diagnostic.
const maxRawBody = 512

type APIError struct {
	StatusCode int
	Raw        string
	Render     *apitypes.RenderErrorResponse
}

func ErrorFromResponse(resp *http.Response, body []byte) *APIError {
	if resp == nil {
		return &APIError{Raw: rawBody(body)}
	}

	return decodeError(resp.StatusCode, body)
}

func (e *APIError) Error() string {
	if e.Render != nil {
		if pretty, err := json.MarshalIndent(e.Render, "", "  "); err == nil {
			return fmt.Sprintf("signoz api: HTTP %d\n%s", e.StatusCode, pretty)
		}
	}

	if e.Raw != "" {
		return fmt.Sprintf("signoz api: HTTP %d: %s", e.StatusCode, e.Raw)
	}

	return fmt.Sprintf("signoz api: HTTP %d", e.StatusCode)
}

func decodeError(status int, raw []byte) *APIError {
	out := &APIError{StatusCode: status, Raw: rawBody(raw)}

	var re apitypes.RenderErrorResponse
	if err := json.Unmarshal(raw, &re); err != nil {
		return out
	}

	if re.Error.Code != "" || re.Error.Message != "" || (re.Error.Errors != nil && len(*re.Error.Errors) > 0) {
		out.Render = &re
	}

	return out
}

// rawBody trims surrounding whitespace and caps the body at maxRawBody bytes so
// an oversized non-envelope response (e.g. a proxy error page) doesn't bloat the
// diagnostic.
func rawBody(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > maxRawBody {
		return s[:maxRawBody] + "… (truncated)"
	}

	return s
}
