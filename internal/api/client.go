package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/smoolzone/odux-agent/internal/config"
)

type Client struct {
	Base  string
	Token string
	HTTP  *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		Base:  cfg.Hub,
		Token: cfg.Token,
		HTTP:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) do(method, path string, body any, out any) error {
	return request(c.Base, c.Token, method, path, body, out)
}

// request performs a JSON request. token == "" sends no Authorization header.
func request(base, token, method, path string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, base+path, rdr)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach hub %s: %w", base, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hub returned %s: %s", resp.Status, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

type Server struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// Servers lists the workspace's chambers, optionally filtered by a label selector
// (e.g. "role=db,region=eu"). Uses the panel endpoint (/api/servers) with a session
// access token — the /api/v1/servers route is the public API-key surface, not this.
func (c *Client) Servers(accessToken, selector string) ([]Server, error) {
	path := "/api/servers"
	if selector != "" {
		path += "?selector=" + url.QueryEscape(selector)
	}
	var out []Server
	if err := request(c.Base, accessToken, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Server(id string) (*Server, error) {
	var out Server
	err := c.do("GET", "/api/v1/servers/"+id, nil, &out)
	return &out, err
}

// ServerByToken fetches one chamber via the panel endpoint with a session token.
func (c *Client) ServerByToken(accessToken, id string) (*Server, error) {
	var out Server
	if err := request(c.Base, accessToken, "GET", "/api/servers/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveServer resolves a chamber reference (numeric id, name, or host) to its id.
func (c *Client) ResolveServer(accessToken, ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("chamber reference is required")
	}
	if _, err := strconv.Atoi(ref); err == nil {
		return ref, nil // already a numeric id
	}
	servers, err := c.Servers(accessToken, "")
	if err != nil {
		return "", err
	}
	for _, s := range servers {
		if s.Name == ref || s.Host == ref {
			return strconv.Itoa(s.ID), nil
		}
	}
	return "", fmt.Errorf("chamber %q not found (use its id, name, or host)", ref)
}

// Login authenticates a box operator (used only by `odux box setup`; the
// password is consumed in memory and never persisted).
func Login(hub, email, password string) (string, error) {
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	payload := map[string]string{"email": email, "password": password}
	if err := request(hub, "", "POST", "/api/auth/login", payload, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("login succeeded but returned no access token")
	}
	return out.AccessToken, nil
}

// CreateWorkspaceAPIKey creates a workspace API key (used by `odux box setup`).
func CreateWorkspaceAPIKey(hub, jwt, name string) (string, error) {
	var out struct {
		Secret string `json:"secret"`
	}
	payload := struct {
		Name string `json:"name"`
	}{Name: name}
	if err := request(hub, jwt, "POST", "/api/workspace-api-keys", payload, &out); err != nil {
		return "", err
	}
	if out.Secret == "" {
		return "", fmt.Errorf("key creation returned no secret")
	}
	return out.Secret, nil
}

// TokenPair is the login/refresh response. Only the refresh token is ever
// persisted (encrypted); the access token lives in memory for one command.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// LoginPair authenticates with email + password (+ optional OTP) and returns
// both tokens. Used by `odux login`.
func LoginPair(hub, email, password, otp string) (*TokenPair, error) {
	var out TokenPair
	payload := map[string]string{"email": email, "password": password}
	if otp != "" {
		payload["otp"] = otp
	}
	if err := request(hub, "", "POST", "/api/auth/login", payload, &out); err != nil {
		return nil, err
	}
	if out.RefreshToken == "" {
		return nil, fmt.Errorf("login succeeded but returned no refresh token")
	}
	return &out, nil
}

// Refresh exchanges a refresh token for a fresh short-lived access token.
func Refresh(hub, refreshToken string) (string, error) {
	var out struct {
		AccessToken string `json:"accessToken"`
	}
	if err := request(hub, "", "POST", "/api/auth/refresh", map[string]string{"refreshToken": refreshToken}, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("refresh returned no access token")
	}
	return out.AccessToken, nil
}

// RefreshPair exchanges a refresh token for a fresh pair. The server rotates the
// refresh token on every use (old is revoked), so callers MUST persist the new
// refreshToken returned here or the next command will 401 with INVALID_REFRESH.
func RefreshPair(hub, refreshToken string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("this command needs a login session — run `odux login`")
	}
	var out TokenPair
	if err := request(hub, "", "POST", "/api/auth/refresh", map[string]string{"refreshToken": refreshToken}, &out); err != nil {
		return nil, err
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		return nil, fmt.Errorf("refresh returned an incomplete token pair")
	}
	return &out, nil
}

// Logout revokes the refresh token server-side.
func Logout(hub, refreshToken string) error {
	return request(hub, refreshToken, "POST", "/api/auth/logout", map[string]string{"refreshToken": refreshToken}, nil)
}

// AccessToken resolves the access token for this command: a session refreshes
// itself; a read-only workspace key cannot be used for writes.
func (c *Client) AccessToken(refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("this command needs a login session — run `odux login`")
	}
	return Refresh(c.Base, refreshToken)
}

type ServerInput struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Role     string `json:"role,omitempty"`
	Group    string `json:"group,omitempty"`
	Tags     string `json:"tags,omitempty"`
	SSHUser  string `json:"sshUser,omitempty"`
	SSHPort  int    `json:"sshPort,omitempty"`
	SSHKeyID int    `json:"sshKeyId,omitempty"`
}

func (c *Client) CreateServer(accessToken string, in ServerInput) (*Server, error) {
	var out Server
	if err := request(c.Base, accessToken, "POST", "/api/servers", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteServer(accessToken, id string) error {
	return request(c.Base, accessToken, "DELETE", "/api/servers/"+id, nil, nil)
}

type DeploymentInput struct {
	ServerID        int    `json:"serverId"`
	Repo            string `json:"repo"`
	Branch          string `json:"branch,omitempty"`
	BuildCommand    string `json:"buildCommand"`
	RollbackCommand string `json:"rollbackCommand,omitempty"`
}

func (c *Client) CreateDeployment(accessToken string, in DeploymentInput) (int, error) {
	var out struct {
		ID int `json:"id"`
	}
	if err := request(c.Base, accessToken, "POST", "/api/ops/deployments", in, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

func (c *Client) ApproveDeployment(accessToken, id string) error {
	return request(c.Base, accessToken, "POST", "/api/ops/deployments/"+id+"/approve", nil, nil)
}

func (c *Client) RunDeployment(accessToken, id string) error {
	return request(c.Base, accessToken, "POST", "/api/ops/deployments/"+id+"/run", nil, nil)
}

func (c *Client) RollbackDeployment(accessToken, id string) error {
	return request(c.Base, accessToken, "POST", "/api/ops/deployments/"+id+"/rollback", nil, nil)
}

type WorkspaceInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Me struct {
	ID         int             `json:"id"`
	Email      string          `json:"email"`
	Fullname   string          `json:"fullname"`
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

func (c *Client) Me(accessToken string) (*Me, error) {
	var out Me
	if err := request(c.Base, accessToken, "GET", "/api/auth/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type Member struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (c *Client) WorkspaceMembers(accessToken string) ([]Member, error) {
	var out []Member
	if err := request(c.Base, accessToken, "GET", "/api/workspace/members", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get performs a GET with the given access token.
func (c *Client) Get(path, token string, out any) error {
	return request(c.Base, token, "GET", path, nil, out)
}

// Write performs a request with the given access token and optional body.
func (c *Client) Write(method, path, token string, body, out any) error {
	return request(c.Base, token, method, path, body, out)
}
