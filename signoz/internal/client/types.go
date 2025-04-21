package client

import "github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"

// signozResponse - Maps the response data.
type signozResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data"`
	ErrorType string      `json:"errorType"`
	Error     string      `json:"error"`
}

// alertResponse - Maps the response data of GetAlert and CreateAlert.
type alertResponse struct {
	Status    string      `json:"status"`
	Error     string      `json:"error"`
	ErrorType string      `json:"errorType"`
	Data      model.Alert `json:"data"`
}

// dashboardRespose - Maps the response data of
type dashboardResponse struct {
	Status    string          `json:"status"`
	Error     string          `json:"error"`
	ErrorType string          `json:"errorType"`
	Data      model.Dashboard `json:"data"`
}
