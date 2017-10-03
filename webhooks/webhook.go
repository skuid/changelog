package webhooks

import (
	"net/http"
)

const WebhookContextPullRequest = "changelog/pull-request"
const WebhookContextPush = "changelog/push"

type VCSWebhook interface {
	New(secret string, apiToken string) http.Handler
}
