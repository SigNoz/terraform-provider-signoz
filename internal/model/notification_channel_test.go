package model

import (
	"encoding/json"
	"testing"
)

func TestNotificationChannel_SlackJSONRoundTrip(t *testing.T) {
	ch := NotificationChannel{
		ID:   "1",
		Name: "slack-alerts",
		Type: ChannelTypeSlack,
		SlackConfigs: []SlackConfig{
			{
				ApiUrl:  "https://hooks.slack.com/services/T00/B00/XXX",
				Channel: "#alerts",
				Title:   "Alert Fired",
				Text:    "Something happened",
			},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if ch2.ID != ch.ID {
		t.Errorf("ID mismatch: %s != %s", ch2.ID, ch.ID)
	}
	if ch2.Name != ch.Name {
		t.Errorf("Name mismatch: %s != %s", ch2.Name, ch.Name)
	}
	if ch2.Type != ChannelTypeSlack {
		t.Errorf("Type mismatch: %s != %s", ch2.Type, ChannelTypeSlack)
	}
	if len(ch2.SlackConfigs) != 1 {
		t.Fatalf("expected 1 slack config, got %d", len(ch2.SlackConfigs))
	}
	if ch2.SlackConfigs[0].ApiUrl != ch.SlackConfigs[0].ApiUrl {
		t.Errorf("ApiUrl mismatch")
	}
	if ch2.SlackConfigs[0].Channel != "#alerts" {
		t.Errorf("Channel mismatch: %s", ch2.SlackConfigs[0].Channel)
	}
}

func TestNotificationChannel_WebhookJSONRoundTrip(t *testing.T) {
	ch := NotificationChannel{
		ID:   "2",
		Name: "webhook-alerts",
		Type: ChannelTypeWebhook,
		WebhookConfigs: []WebhookConfig{
			{
				ApiUrl:   "https://example.com/webhook",
				Username: "user",
				Password: "pass",
			},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(ch2.WebhookConfigs) != 1 {
		t.Fatalf("expected 1 webhook config, got %d", len(ch2.WebhookConfigs))
	}
	if ch2.WebhookConfigs[0].Username != "user" {
		t.Errorf("Username mismatch: %s", ch2.WebhookConfigs[0].Username)
	}
	if ch2.WebhookConfigs[0].Password != "pass" {
		t.Errorf("Password mismatch: %s", ch2.WebhookConfigs[0].Password)
	}
}

func TestNotificationChannel_PagerdutyJSONRoundTrip(t *testing.T) {
	ch := NotificationChannel{
		ID:   "3",
		Name: "pagerduty-alerts",
		Type: ChannelTypePagerduty,
		PagerdutyConfigs: []PagerdutyConfig{
			{
				RoutingKey: "routing-key-123",
				ServiceKey: "service-key-456",
				Details: map[string]string{
					"env":     "production",
					"service": "api",
				},
			},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(ch2.PagerdutyConfigs) != 1 {
		t.Fatalf("expected 1 pagerduty config, got %d", len(ch2.PagerdutyConfigs))
	}
	cfg := ch2.PagerdutyConfigs[0]
	if cfg.RoutingKey != "routing-key-123" {
		t.Errorf("RoutingKey mismatch: %s", cfg.RoutingKey)
	}
	if cfg.Details["env"] != "production" {
		t.Errorf("Details[env] mismatch: %s", cfg.Details["env"])
	}
}

func TestNotificationChannel_OpsgenieJSONRoundTrip(t *testing.T) {
	ch := NotificationChannel{
		ID:   "4",
		Name: "opsgenie-alerts",
		Type: ChannelTypeOpsgenie,
		OpsgenieConfigs: []OpsgenieConfig{
			{
				ApiKey: "api-key-789",
				ApiUrl: "https://api.opsgenie.com/v2/alerts",
			},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(ch2.OpsgenieConfigs) != 1 {
		t.Fatalf("expected 1 opsgenie config, got %d", len(ch2.OpsgenieConfigs))
	}
	if ch2.OpsgenieConfigs[0].ApiKey != "api-key-789" {
		t.Errorf("ApiKey mismatch: %s", ch2.OpsgenieConfigs[0].ApiKey)
	}
}

func TestNotificationChannel_MSTeamsJSONRoundTrip(t *testing.T) {
	ch := NotificationChannel{
		ID:   "5",
		Name: "teams-alerts",
		Type: ChannelTypeMSTeams,
		MSTeamsConfigs: []MSTeamsConfig{
			{
				WebhookUrl: "https://outlook.office.com/webhook/xxx",
			},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(ch2.MSTeamsConfigs) != 1 {
		t.Fatalf("expected 1 msteams config, got %d", len(ch2.MSTeamsConfigs))
	}
	if ch2.MSTeamsConfigs[0].WebhookUrl != "https://outlook.office.com/webhook/xxx" {
		t.Errorf("WebhookUrl mismatch: %s", ch2.MSTeamsConfigs[0].WebhookUrl)
	}
}

func TestNotificationChannel_EmailJSONRoundTrip(t *testing.T) {
	ch := NotificationChannel{
		ID:   "6",
		Name: "email-alerts",
		Type: ChannelTypeEmail,
		EmailConfigs: []EmailConfig{
			{
				SendResolved: true,
				To:           "oncall@example.com",
				Html:         "<h1>Alert</h1>",
				Headers: map[string]string{
					"X-Priority": "1",
				},
			},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(ch2.EmailConfigs) != 1 {
		t.Fatalf("expected 1 email config, got %d", len(ch2.EmailConfigs))
	}
	cfg := ch2.EmailConfigs[0]
	if !cfg.SendResolved {
		t.Error("expected SendResolved=true")
	}
	if cfg.To != "oncall@example.com" {
		t.Errorf("To mismatch: %s", cfg.To)
	}
	if cfg.Headers["X-Priority"] != "1" {
		t.Errorf("Headers[X-Priority] mismatch: %s", cfg.Headers["X-Priority"])
	}
}

func TestNotificationChannel_EmptyConfigs(t *testing.T) {
	ch := NotificationChannel{
		ID:   "10",
		Name: "empty",
		Type: ChannelTypeSlack,
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if ch2.SlackConfigs != nil {
		t.Errorf("expected nil SlackConfigs with omitempty, got %v", ch2.SlackConfigs)
	}
	if ch2.WebhookConfigs != nil {
		t.Errorf("expected nil WebhookConfigs, got %v", ch2.WebhookConfigs)
	}
}

func TestNotificationChannel_ChannelTypes(t *testing.T) {
	expected := []string{"slack", "webhook", "pagerduty", "opsgenie", "msteams", "email"}
	if len(ChannelTypes) != len(expected) {
		t.Fatalf("expected %d channel types, got %d", len(expected), len(ChannelTypes))
	}
	for i, ct := range expected {
		if ChannelTypes[i] != ct {
			t.Errorf("channel type[%d]: expected %s, got %s", i, ct, ChannelTypes[i])
		}
	}
}

func TestNotificationChannel_MultipleConfigs(t *testing.T) {
	ch := NotificationChannel{
		ID:   "7",
		Name: "multi-slack",
		Type: ChannelTypeSlack,
		SlackConfigs: []SlackConfig{
			{ApiUrl: "https://hooks.slack.com/1", Channel: "#alerts"},
			{ApiUrl: "https://hooks.slack.com/2", Channel: "#critical"},
		},
	}

	data, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var ch2 NotificationChannel
	err = json.Unmarshal(data, &ch2)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(ch2.SlackConfigs) != 2 {
		t.Fatalf("expected 2 slack configs, got %d", len(ch2.SlackConfigs))
	}
}
