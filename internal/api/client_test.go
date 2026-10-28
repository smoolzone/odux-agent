package api

import (
	"strings"
	"testing"
)

func TestCheckHubRequiresTLSExceptOnThisMachine(t *testing.T) {
	for _, ok := range []string{
		"https://app.odux.io",
		"https://hub.example.com:8443",
		"http://localhost:3200",
		"http://127.0.0.1:3200",
		"http://[::1]:3200",
	} {
		if err := CheckHub(ok); err != nil {
			t.Errorf("%s: expected accepted, got %v", ok, err)
		}
	}
	for _, bad := range []string{
		"http://app.odux.io",
		"http://10.0.0.5:3200",
		"ftp://app.odux.io",
		"app.odux.io",
		"",
	} {
		if err := CheckHub(bad); err == nil {
			t.Errorf("%s: expected refused", bad)
		}
	}
	if err := CheckHub("http://app.odux.io"); !strings.Contains(err.Error(), "unencrypted") {
		t.Errorf("the refusal should say why: %v", err)
	}
}

func TestEveryRequestIsCheckedAndHasATimeout(t *testing.T) {
	if httpClient.Timeout == 0 {
		t.Fatal("hub calls need a timeout, or a stalled hub hangs the agent")
	}
	err := request("http://app.odux.io", "", "GET", "/api/x", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "refusing plain http") {
		t.Fatalf("a plain-http hub must be refused before anything is sent, got %v", err)
	}
}
