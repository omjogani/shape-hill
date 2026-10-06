package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/omjogani/shape-hill/api/internal/account"
	"github.com/omjogani/shape-hill/api/internal/hills"
)

type apiError struct {
	Status  int
	Message string
}

func (e *apiError) Error() string { return e.Message }

type transportError struct{ err error }

func (e *transportError) Error() string { return e.err.Error() }

type client struct {
	base  string
	token string
	http  *http.Client
}

func newClient(base, token string) *client {
	return &client{
		base:  strings.TrimRight(base, "/"),
		token: token,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *client) do(ctx context.Context, method, path string, body, out any) error {
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return &transportError{fmt.Errorf("could not reach %s: %w", c.base, err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var problem struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&problem)
		if problem.Error == "" {
			problem.Error = resp.Status
		}
		return &apiError{Status: resp.StatusCode, Message: problem.Error}
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return &transportError{fmt.Errorf("unexpected response from %s: %w", c.base, err)}
	}
	return nil
}

type hillWithScopes struct {
	Hill   hills.Hill    `json:"hill"`
	Scopes []hills.Scope `json:"scopes"`
}

func (c *client) me(ctx context.Context) (account.User, error) {
	var body struct {
		Onboarded bool         `json:"onboarded"`
		User      account.User `json:"user"`
	}
	err := c.do(ctx, http.MethodGet, "/api/me", nil, &body)
	return body.User, err
}

func (c *client) listHills(ctx context.Context) ([]hills.Hill, error) {
	var owned []hills.Hill
	err := c.do(ctx, http.MethodGet, "/api/hills", nil, &owned)
	return owned, err
}

func (c *client) createHill(ctx context.Context, slug, title, description string, isPublic bool) (hills.Hill, error) {
	var hill hills.Hill
	err := c.do(ctx, http.MethodPost, "/api/hills", map[string]any{
		"slug": slug, "title": title, "description": description, "is_public": isPublic,
	}, &hill)
	return hill, err
}

func (c *client) hill(ctx context.Context, slug string) (hillWithScopes, error) {
	var found hillWithScopes
	err := c.do(ctx, http.MethodGet, "/api/hills/"+url.PathEscape(slug), nil, &found)
	return found, err
}

func (c *client) addScope(ctx context.Context, slug, title, description, color string, sortOrder int) (hills.Scope, error) {
	var scope hills.Scope
	err := c.do(ctx, http.MethodPost, "/api/hills/"+url.PathEscape(slug)+"/scopes", map[string]any{
		"title": title, "description": description, "color": color, "sort_order": sortOrder,
	}, &scope)
	return scope, err
}

func (c *client) moveScope(ctx context.Context, scopeID string, position int16, note string) error {
	return c.do(ctx, http.MethodPost, "/api/scopes/"+url.PathEscape(scopeID)+"/positions", map[string]any{
		"position": position, "note": note,
	}, nil)
}

func (c *client) snapshots(ctx context.Context, scopeID string) ([]hills.Snapshot, error) {
	var snaps []hills.Snapshot
	err := c.do(ctx, http.MethodGet, "/api/scopes/"+url.PathEscape(scopeID)+"/positions", nil, &snaps)
	return snaps, err
}
