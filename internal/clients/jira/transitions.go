package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"buddy/internal/errors"
)

// CloseTicket closes a JIRA ticket by transitioning it
func (c *JiraClient) CloseTicket(ctx context.Context, issueKey string, reasonType string) error {
	transitions, err := c.getAvailableTransitions(ctx, issueKey)
	if err != nil {
		return err
	}

	targetTransitionID := c.findCloseTransition(transitions)
	if targetTransitionID == "" {
		return errors.NotFound("close transition")
	}

	return c.executeTransition(ctx, issueKey, targetTransitionID)
}

// getAvailableTransitions fetches available transitions for an issue
func (c *JiraClient) getAvailableTransitions(ctx context.Context, issueKey string) ([]transitionInfo, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.config.Domain, issueKey)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to create transitions request")
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "transitions request failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var transitions struct {
		Transitions []transitionInfo `json:"transitions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&transitions); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to decode transitions")
	}

	return transitions.Transitions, nil
}

// findCloseTransition finds a transition that closes the issue
func (c *JiraClient) findCloseTransition(transitions []transitionInfo) string {
	for _, t := range transitions {
		lowerName := strings.ToLower(t.Name)
		if strings.Contains(lowerName, "done") || strings.Contains(lowerName, "close") {
			return t.ID
		}
	}
	return ""
}

// executeTransition executes a transition on an issue
func (c *JiraClient) executeTransition(ctx context.Context, issueKey, transitionID string) error {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", c.config.Domain, issueKey)

	transitionReq := struct {
		Transition struct {
			ID string `json:"id"`
		} `json:"transition"`
	}{}
	transitionReq.Transition.ID = transitionID

	reqBody, err := json.Marshal(transitionReq)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to marshal transition request")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExternal, "failed to create transition request")
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExternal, "transition failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleAPIError(resp)
	}

	c.logger.Info("Successfully closed ticket: %s", issueKey)
	return nil
}

// transitionInfo represents a JIRA transition
type transitionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
