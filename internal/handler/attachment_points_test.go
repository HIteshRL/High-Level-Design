package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAttachmentCatalog(t *testing.T) {
	h := NewAttachmentHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/architecture/attachment-points", nil)

	h.Catalog(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "image-inference") {
		t.Fatalf("expected catalog to include image-inference domain")
	}
}

func TestSubmitImageJob(t *testing.T) {
	h := NewAttachmentHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image/jobs", nil)

	h.SubmitImageJob(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "image-jobs") {
		t.Fatalf("expected image queue name in response")
	}
}

func TestNotImplemented(t *testing.T) {
	h := NewAttachmentHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rag/search", nil)

	h.NotImplemented(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected status 501, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "planned") {
		t.Fatalf("expected planned status in response")
	}
}
