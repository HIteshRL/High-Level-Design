// Attachment point handler surfaces non-essential subsystem API contracts.
// Maps to design.swift modules that are currently planned but not fully implemented.
package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prakyathpnayak/roognis/internal/models"
)

// AttachmentHandler exposes low-level design entry points for future modules.
type AttachmentHandler struct{}

func NewAttachmentHandler() *AttachmentHandler {
	return &AttachmentHandler{}
}

// Catalog returns all non-essential API attachment points.
func (h *AttachmentHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	items := []models.AttachmentPoint{
		{Domain: "image-inference", Feature: "submit-job", Method: "POST", Path: "/api/v1/image/jobs", Description: "Queue image generation jobs", DependsOn: []string{"Kafka/SQS", "GPU Allocation Manager", "Artifact Storage"}},
		{Domain: "image-inference", Feature: "job-status", Method: "GET", Path: "/api/v1/image/jobs/{id}", Description: "Retrieve queued image job status"},
		{Domain: "video-inference", Feature: "submit-job", Method: "POST", Path: "/api/v1/video/jobs", Description: "Queue video/simulation generation jobs", DependsOn: []string{"Priority Queue", "GPU Pool Manager", "Checkpointing"}},
		{Domain: "video-inference", Feature: "job-status", Method: "GET", Path: "/api/v1/video/jobs/{id}", Description: "Retrieve queued video job status"},
		{Domain: "rag", Feature: "upsert-documents", Method: "POST", Path: "/api/v1/rag/documents", Description: "Ingest and normalize source documents"},
		{Domain: "rag", Feature: "search", Method: "POST", Path: "/api/v1/rag/search", Description: "Retrieve and rerank context candidates"},
		{Domain: "psychographic", Feature: "ingest-events", Method: "POST", Path: "/api/v1/psychographic/events", Description: "Ingest behavioral events for profile updates"},
		{Domain: "psychographic", Feature: "persona-state", Method: "GET", Path: "/api/v1/psychographic/persona/{user_id}", Description: "Return persona summary used by orchestration"},
		{Domain: "quiz", Feature: "generate", Method: "POST", Path: "/api/v1/quiz/generate", Description: "Generate adaptive concept-based questions"},
		{Domain: "quiz", Feature: "submit-attempt", Method: "POST", Path: "/api/v1/quiz/attempts", Description: "Persist learner quiz attempt and score"},
		{Domain: "analytics", Feature: "kpi-report", Method: "GET", Path: "/api/v1/analytics/kpi", Description: "Expose ops and pedagogy KPI aggregates"},
	}

	writeJSON(w, http.StatusOK, models.AttachmentCatalogResponse{
		GeneratedAt: time.Now().UTC(),
		Version:     "v1",
		Items:       items,
	})
}

func (h *AttachmentHandler) SubmitImageJob(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusAccepted, models.AsyncJobAccepted{
		JobID:       uuid.NewString(),
		Status:      "accepted",
		SubmittedAt: time.Now().UTC(),
		Queue:       "image-jobs",
	})
}

func (h *AttachmentHandler) SubmitVideoJob(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusAccepted, models.AsyncJobAccepted{
		JobID:       uuid.NewString(),
		Status:      "accepted",
		SubmittedAt: time.Now().UTC(),
		Queue:       "video-jobs",
	})
}

func (h *AttachmentHandler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, models.NotImplementedResponse{
		Capability:   r.Method + " " + r.URL.Path,
		Status:       "planned",
		RoadmapPhase: "non-essential-architecture",
		Owner:        "platform",
		SuggestedNext: []string{
			"Define domain request/response schemas",
			"Wire service interface and persistence implementation",
			"Add integration tests for happy-path and failure-path",
		},
	})
}
