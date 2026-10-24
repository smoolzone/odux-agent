package api

import "fmt"

// Policy is a versioned desired-state bundle (metadata only).
type Policy struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Version     int    `json:"version"`
	Selector    string `json:"selector"`
	Status      string `json:"status"`
	RequestedBy *int   `json:"requestedById"`
	ApprovedBy  *int   `json:"approvedById"`
	CreatedAt   string `json:"createdAt"`
}

// PolicyDetail adds the parsed spec.
type PolicyDetail struct {
	Policy
	Spec map[string]any `json:"spec"`
}

// ListPolicies returns all policies (newest version first per name).
func (c *Client) ListPolicies(token string) ([]Policy, error) {
	var out []Policy
	if err := request(c.Base, token, "GET", "/api/policies", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreatePolicy creates a draft v1.
func (c *Client) CreatePolicy(token, name, selector string, spec map[string]any) (*Policy, error) {
	var out Policy
	payload := map[string]any{"name": name, "selector": selector, "spec": spec}
	if err := request(c.Base, token, "POST", "/api/policies", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPolicy returns a policy with its spec.
func (c *Client) GetPolicy(token, id string) (*PolicyDetail, error) {
	var out PolicyDetail
	if err := request(c.Base, token, "GET", "/api/policies/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PolicyAction performs a lifecycle transition (submit/approve/reject/activate/archive).
func (c *Client) PolicyAction(token, id, action string) error {
	return request(c.Base, token, "POST", fmt.Sprintf("/api/policies/%s/%s", id, action), nil, nil)
}

// PolicyTargets resolves the chambers a policy targets.
func (c *Client) PolicyTargets(token, id string) ([]Server, error) {
	var out []Server
	if err := request(c.Base, token, "GET", "/api/policies/"+id+"/targets", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PolicyApplyResult reports what a policy apply enqueued vs. skipped.
type PolicyApplyResult struct {
	Enqueued []struct {
		ServerID  int    `json:"serverId"`
		Name      string `json:"name"`
		CommandID int    `json:"commandId"`
	} `json:"enqueued"`
	Skipped []struct {
		ServerID int    `json:"serverId"`
		Name     string `json:"name"`
		Reason   string `json:"reason"`
	} `json:"skipped"`
}

// PolicyApply enqueues deploy commands on the policy's target chambers.
func (c *Client) PolicyApply(token, id string) (*PolicyApplyResult, error) {
	var out PolicyApplyResult
	if err := request(c.Base, token, "POST", "/api/policies/"+id+"/apply", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
