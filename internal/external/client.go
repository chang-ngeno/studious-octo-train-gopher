package external

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type TokenTransport struct {
	AuthURL, User, Pass string
	token               string
	expiry              time.Time
	mu                  sync.RWMutex
	Underlying          http.RoundTripper
	BaseClient          *http.Client
}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.RLock()
	expired := t.token == "" || time.Now().Add(30*time.Second).After(t.expiry)
	t.mu.RUnlock()

	if expired {
		if err := t.refresh(); err != nil { return nil, err }
	}

	t.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+t.token)
	t.mu.RUnlock()

	return t.Underlying.RoundTrip(req)
}

func (t *TokenTransport) refresh() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var lastErr error
	for i := 0; i < 3; i++ { // Retry logic
		form := url.Values{"grant_type": {"client_credentials"}}
		req, _ := http.NewRequest("POST", t.AuthURL, strings.NewReader(form.Encode()))
		req.SetBasicAuth(t.User, t.Pass)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := t.BaseClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			var res struct { AccessToken string `json:"access_token"`; ExpiresIn int `json:"expires_in"` }
			json.NewDecoder(resp.Body).Decode(&res)
			t.token = res.AccessToken
			t.expiry = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
			resp.Body.Close()
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(1<<i) * time.Second) // Exponential backoff
	}
	return lastErr
}


