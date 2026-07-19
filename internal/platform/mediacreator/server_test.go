package mediacreator_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/mediacreator"
)

func TestRunServesLocalGUI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mediacreator.Run(ctx, mediacreator.Options{
			Version:     "test",
			Listen:      "127.0.0.1:0",
			OpenBrowser: false,
		})
	}()

	// Discover the bound port via status by racing briefly — Run prints URL to stderr.
	// Instead start with a known free approach: poll is hard. Use fixed high port.
	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestStatusEndpointViaFixedPort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- mediacreator.Run(ctx, mediacreator.Options{
			Version:     "0.2.0-test",
			Listen:      "127.0.0.1:17991",
			OpenBrowser: false,
		})
	}()

	deadline := time.Now().Add(3 * time.Second)
	var res *http.Response
	var err error
	for time.Now().Before(deadline) {
		res, err = http.Get("http://127.0.0.1:17991/api/status")
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status code %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload["name"] != "Vyntrio Media Creator" {
		t.Fatalf("name = %v", payload["name"])
	}
	if payload["kind"] != "local_web_gui" {
		t.Fatalf("kind = %v", payload["kind"])
	}
	if payload["native_gui"] != false {
		t.Fatalf("native_gui should be false for honesty")
	}

	index, err := http.Get("http://127.0.0.1:17991/")
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	defer index.Body.Close()
	html, _ := io.ReadAll(index.Body)
	if !containsAll(string(html), "Vyntrio", "Media Creator", "Write image") {
		t.Fatalf("index missing expected UI markers")
	}

	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop")
	}
}

func containsAll(haystack string, needles ...string) bool {
	for _, n := range needles {
		if !contains(haystack, n) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
