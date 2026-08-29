package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestTheRootPathIsServed keeps the request surface a real artefact rather than
// an empty package the invariant in `internal/zonesync/offrequestpath_test.go`
// happens to find. A seam that serves nothing is a seam nobody would notice
// deleting.
func TestTheRootPathIsServed(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	NewMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.Len() == 0 {
		t.Error("the body is empty, want a non-empty response")
	}
}
