package text

import (
	"bytes"
	"testing"

	"github.com/azargarov/rsvpck/internal/domain"
)

func TestTruncateErrorASCII(t *testing.T) {
	long := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	got := truncateErrorASCII(long, 10)
	if got != "aaaaaaa..." {
		t.Fatalf("truncate ASCII got %q", got)
	}
}

func TestTruncateErrorUTF8(t *testing.T) {
	long := "Hello world" + "verylongdash"
	got := truncateError(long, 8)
	// should not split runes; last three dots reserved if >=3
	if got[len(got)-3:] != "..." {
		t.Fatalf("want ellipsis, got %q", got)
	}
}

func TestPrintSummary_StableASCII(t *testing.T) {
	var buf bytes.Buffer
	rc := NewRenderConfig(WithForceASCII(true))
	res := domain.NewConnectivityResult(domain.ModeViaProxy, nil)
	printSummary(&buf, res, rc)
	s := buf.String()
	if rc.Unicode {
		t.Fatalf("expected ASCII mode")
	}
	if s == "" {
		t.Fatalf("empty summary")
	}
}
