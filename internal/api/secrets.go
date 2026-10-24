package api

// Secret is a workspace secret's metadata (no value).
type Secret struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Note      string `json:"note"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// SecretValue is a secret with its decrypted value (explicit read only).
type SecretValue struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

// ListSecrets returns secret metadata (values are never returned).
func (c *Client) ListSecrets(token string) ([]Secret, error) {
	var out []Secret
	if err := request(c.Base, token, "GET", "/api/secrets", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetSecret creates or overwrites a secret. The value is encrypted at rest.
func (c *Client) SetSecret(token, key, value, note string) (*Secret, error) {
	var out Secret
	payload := map[string]string{"key": key, "value": value, "note": note}
	if err := request(c.Base, token, "POST", "/api/secrets", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSecret reads (and decrypts) a single secret by id.
func (c *Client) GetSecret(token, id string) (*SecretValue, error) {
	var out SecretValue
	if err := request(c.Base, token, "GET", "/api/secrets/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSecret removes a secret by id.
func (c *Client) DeleteSecret(token, id string) error {
	return request(c.Base, token, "DELETE", "/api/secrets/"+id, nil, nil)
}
