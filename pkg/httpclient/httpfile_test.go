package httpclient

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltinHTTPDownload(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := builtinHTTPDownload(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, scytheDownloadHeader) {
		t.Fatalf("unexpected header: %s", out[:min(80, len(out))])
	}
	if !strings.Contains(out, "encoding=base64") {
		t.Fatal("missing base64 marker")
	}
}

func TestHandleHTTPFileCommandMap_DownloadShorthand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "x.bin")
	_ = os.WriteFile(p, []byte{1, 2, 3}, 0644)

	out, err := handleHTTPFileCommandMap(map[string]interface{}{"download": p})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "encoding=base64") {
		t.Fatal(out)
	}
}

func TestSummarizeHTTPFileCommandForReceive_StripsBase64(t *testing.T) {
	t.Parallel()
	m := map[string]interface{}{
		"op":               "upload",
		"path":             "/tmp/x",
		"content_base64":   strings.Repeat("YQ", 5000),
		"contentBase64":    "skip-me",
	}
	s := summarizeHTTPFileCommandForReceive(m)
	if strings.Contains(s, "content_base64") || strings.Contains(s, "contentBase64") || strings.Contains(s, "YQ") {
		t.Fatalf("summary should omit payload keys: %s", s[:min(200, len(s))])
	}
	if !strings.Contains(s, "/tmp/x") {
		t.Fatal(s)
	}
}

func TestExecuteStructuredBeaconCommand_SummaryOmitsBase64(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.dat")
	large := strings.Repeat("YQ", 100) // valid base64 pairs
	sum, out, err := executeStructuredBeaconCommand(map[string]interface{}{
		"path":           dest,
		"content_base64": large,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sum, "content_base64") || strings.Contains(sum, large[:40]) {
		t.Fatalf("bad summary: %s", sum)
	}
	b, _ := os.ReadFile(dest)
	if len(b) == 0 {
		t.Fatal("upload did not run")
	}
	if !strings.Contains(out, "SCYTHE_FILE_UPLOAD") {
		t.Fatal(out)
	}
}

func TestHandleHTTPFileCommandMap_Upload(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.dat")
	_, err := handleHTTPFileCommandMap(map[string]interface{}{
		"path":            dest,
		"content_base64":  "QUJD", // "ABC" in base64
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "ABC" {
		t.Fatalf("got %q", b)
	}
}

func TestRunBuiltinHTTPCommand_DownloadString(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "a")
	_ = os.WriteFile(p, []byte("z"), 0644)
	out, err, handled := runBuiltinHTTPCommand("download " + p)
	if !handled || err != nil {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	if !strings.Contains(out, "encoding=base64") {
		t.Fatal(out)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
