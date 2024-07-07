package client

import "github.com/SigNoz/terraform-provider-signoz/signoz/internal/model"

// signozResponse - Maps the response data.
type signozResponse struct {
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	ErrorType string      `json:"errorType,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// alertResponse - Maps the response data of GetAlert and CreateAlert.
type alertResponse struct {
	Status    string
	Error     string
	ErrorType string
	Data      model.Alert
}
