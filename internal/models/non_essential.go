// Non-essential LLD contracts for pluggable architecture modules.
// These are API attachment-point contracts for future subsystem implementation.
package models

import "time"

// AttachmentPoint describes a not-yet-implemented endpoint with ownership metadata.
type AttachmentPoint struct {
	Domain      string   `json:"domain"`
	Feature     string   `json:"feature"`
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

// AttachmentCatalogResponse lists extension points exposed by the API gateway.
type AttachmentCatalogResponse struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Version     string            `json:"version"`
	Items       []AttachmentPoint `json:"items"`
}

// AsyncJobAccepted is a generic 202 envelope for async image/video pipelines.
type AsyncJobAccepted struct {
	JobID       string    `json:"job_id"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
	Queue       string    `json:"queue"`
}

// NotImplementedResponse includes actionable metadata for stubbed endpoints.
type NotImplementedResponse struct {
	Capability      string   `json:"capability"`
	Status          string   `json:"status"`
	RoadmapPhase    string   `json:"roadmap_phase"`
	Owner           string   `json:"owner"`
	SuggestedNext   []string `json:"suggested_next_steps"`
	TrackingIssueID string   `json:"tracking_issue_id,omitempty"`
}
