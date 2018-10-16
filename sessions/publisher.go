package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bugsnag/bugsnag-go/headers"
)

// sessionPayloadVersion defines the current version of the payload that's
// being sent to the session server.
const sessionPayloadVersion = "1.0"

type sessionPublisher interface {
	publish(sessions []*Session) error
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type publisher struct {
	config *SessionTrackingConfiguration
	client httpClient
}

// publish builds a payload from the given sessions and publishes them to the
// session server. Returns any errors that happened as part of publishing.
func (p *publisher) publish(sessions []*Session) error {
	if p.config.Endpoint == "" {
		// Session tracking is disabled, likely because the notify endpoint was
		// changed without changing the sessions endpoint
		// We've already logged a warning in this case, so no need to spam the
		// log every minute
		return nil
	}
	p.config.mutex.Lock()
	defer p.config.mutex.Unlock()
	payload := makeSessionPayload(sessions, p.config)
	buf, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("bugsnag/sessions/publisher.publish unable to marshal json: %v", err)
	}
	req, err := http.NewRequest("POST", p.config.Endpoint, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("bugsnag/sessions/publisher.publish unable to create request: %v", err)
	}
	for k, v := range headers.PrefixedHeaders(p.config.APIKey, sessionPayloadVersion) {
		req.Header.Add(k, v)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("bugsnag/sessions/publisher.publish unable to deliver session: %v", err)
	}
	defer func(res *http.Response) {
		if err := res.Body.Close(); err != nil {
			p.config.logf("%v", err)
		}
	}(res)
	if res.StatusCode != 202 {
		return fmt.Errorf("bugsnag/session.publish expected 202 response status, got HTTP %s", res.Status)
	}
	return nil
}