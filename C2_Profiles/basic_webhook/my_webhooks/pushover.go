package my_webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

type PushoverPayload struct {
	Title     string `json:"title,omitempty"`
	Message   string `json:"message,omitempty"`
	EventType string `json:"event_type,omitempty"`
	Color     string `json:"color,omitempty"`
}

var slackBoldRegex = regexp.MustCompile(`\*([^*]+)\*`)

func stripSlackMarkdown(s string) string {
	return slackBoldRegex.ReplaceAllString(s, "$1")
}

func sendPushoverMessage(webhookURL string, msg webhookstructs.SlackWebhookMessage) error {
	for _, att := range msg.Attachments {
		payload := PushoverPayload{
			Title:     att.Title,
			Color:     att.Color,
			EventType: inferEventType(att.Title),
		}

		var lines []string

		if att.Blocks != nil {
			for _, block := range *att.Blocks {
				if block.Text != nil && block.Text.Text != "" {
					lines = append(lines, stripSlackMarkdown(block.Text.Text))
				}
				if block.Fields != nil {
					for _, f := range *block.Fields {
						if f.Text != "" {
							lines = append(lines, stripSlackMarkdown(f.Text))
						}
					}
				}
			}
		}

		payload.Message = strings.Join(lines, "\n")

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("pushover basic_webhook returned status %s", resp.Status)
		}
	}

	return nil
}

func inferEventType(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "callback"):
		return "callback"
	case strings.Contains(lower, "alert"):
		return "alert"
	case strings.Contains(lower, "feedback"), strings.Contains(lower, "bug"), strings.Contains(lower, "feature"), strings.Contains(lower, "detection"):
		return "feedback"
	case strings.Contains(lower, "startup"):
		return "startup"
	default:
		return "custom"
	}
}
