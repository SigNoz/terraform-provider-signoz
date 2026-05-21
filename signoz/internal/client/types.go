package client

import (
	"encoding/json"

	"github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"
)

// signozResponse - Maps the response data.
type signozResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	ErrorType string      `json:"errorType"`
	Error     string      `json:"error"`
}

// parseStatusResponse decodes the SigNoz status envelope from a 2xx body.
// Returns parsed=true with the envelope (caller still checks Status/Error);
// parsed=false when the body is empty or not JSON, in which case treat as
// success — doRequest already filtered non-2xx, and some endpoints (notably
// PUT /api/v1/dashboards/{uuid}) return plain-text status rather than the
// envelope. See https://github.com/SigNoz/terraform-provider-signoz/issues/71.
func parseStatusResponse(body []byte) (response signozResponse, parsed bool) {
	if len(body) == 0 {
		return signozResponse{}, false
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return signozResponse{}, false
	}
	return response, true
}

// alertResponse - Maps the response data of GetAlert and CreateAlert.
type alertResponse struct {
	Status    string      `json:"status"`
	Error     string      `json:"error"`
	ErrorType string      `json:"errorType"`
	Data      model.Alert `json:"data"`
}

// dashboardRespose - Maps the response data of CreateDashboard and GetDashboard.
type dashboardResponse struct {
	Status    string        `json:"status"`
	Error     string        `json:"error,omitempty"`
	ErrorType string        `json:"errorType,omitempty"`
	Data      dashboardData `json:"data"`
}

type dashboardData struct {
	CreatedAt string          `json:"createdAt"`
	CreatedBy string          `json:"createdBy"`
	ID        string          `json:"id"`
	Locked    bool            `json:"locked"`
	UpdatedAt string          `json:"updatedAt"`
	UpdatedBy string          `json:"updatedBy"`
	Data      model.Dashboard `json:"data"`
}
