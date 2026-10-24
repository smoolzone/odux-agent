package termio

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadLineVerbatim(t *testing.T) {
	cases := map[string]string{
		"s3cret\n":         "s3cret",
		"with space \r\n":  "with space ",
		"no-newline":       "no-newline",
		"":                 "",
		"  lead+trail  \n": "  lead+trail  ",
		"p@ss'\"$`;|\n":    "p@ss'\"$`;|",
	}
	for in, want := range cases {
		got, err := readLine(strings.NewReader(in))
		if err != nil || got != want {
			t.Errorf("readLine(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

// Runs a tiny program under a real pseudo-terminal (util-linux `script`) and
// checks the typed secret is read correctly but never echoed back.
func TestReadSecretHidesInputOnATerminal(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("uses util-linux script")
	}
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("script(1) not installed")
	}
	bin := filepath.Join(t.TempDir(), "readsecret")
	build := exec.Command("go", "build", "-o", bin, "./testdata/readsecret")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	// Feed the secret through the pty after a short delay so echo is already off.
	cmd := exec.Command("script", "-qfec", "sh -c '"+bin+"'", "/dev/null")
	stdin, _ := cmd.StdinPipe()
	cmd.Env = append(os.Environ(), "TERM=dumb")
	go func() {
		exec.Command("sleep", "0.5").Run()
		stdin.Write([]byte("hunter2-SECRET\n"))
		stdin.Close()
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "GOT=[hunter2-SECRET]") {
		t.Fatalf("secret not read correctly; output:\n%s", s)
	}
	if strings.Count(s, "hunter2-SECRET") != 1 {
		t.Fatalf("secret was echoed to the terminal; output:\n%s", s)
	}
}
