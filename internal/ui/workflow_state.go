package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

// WorkflowState represents the JSON state file for a session's workflow.
type WorkflowState struct {
	Version        int               `json:"version"`
	SessionID      string            `json:"session_id"`
	ClaudeSessionID string           `json:"claude_session_id"`
	Workflow       Workflow          `json:"workflow"`
	Artifacts      []Artifact        `json:"artifacts"`
}

// Workflow tracks pipeline stages and current position.
type Workflow struct {
	Stages  []string `json:"stages"`
	Current string   `json:"current"`
	Status  string   `json:"status"` // active, paused, completed, failed
}

// Artifact types
const (
	ArtifactTypePR      = "pr"       // Pull request (GitHub, Bitbucket, GitLab)
	ArtifactTypeCI      = "ci"       // CI/CD run (Jenkins, GitHub Actions, etc.)
	ArtifactTypeIssue   = "issue"    // Issue tracker (Jira, GitHub Issues, Linear)
	ArtifactTypeCommit  = "commit"   // Git commit
	ArtifactTypeURL     = "url"      // Generic URL (docs, dashboards, etc.)
)

// Artifact represents a tracked output with a well-defined type.
type Artifact struct {
	Type   string `json:"type"`            // pr, ci, issue, commit, url
	Title  string `json:"title"`           // Display name
	URL    string `json:"url,omitempty"`   // Clickable link (required for pr, ci, issue, url)
	Ref    string `json:"ref,omitempty"`   // Short ref (commit SHA, PR number, issue key)
	Status string `json:"status,omitempty"` // open, merged, closed, passed, failed, pending, running
}

// WorkflowWatcher polls a workflow state file for changes.
type WorkflowWatcher struct {
	profile     string
	sessionID   string
	state       *WorkflowState
	stateMu     sync.RWMutex
	lastModTime time.Time
	closeCh     chan struct{}
	closeOnce   sync.Once
	updateCh    chan struct{}
}

// workflowStatePath returns the path to a session's workflow state file.
func workflowStatePath(profile, sessionID string) string {
	dir, err := session.GetProfileDir(profile)
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "workflow", sessionID+".json")
}

// NewWorkflowWatcher creates a watcher for a session's workflow state file.
func NewWorkflowWatcher(profile string) *WorkflowWatcher {
	return &WorkflowWatcher{
		profile:  profile,
		closeCh:  make(chan struct{}),
		updateCh: make(chan struct{}, 1),
	}
}

// SetSession switches which session's workflow file to watch.
func (w *WorkflowWatcher) SetSession(sessionID string) {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()
	if w.sessionID == sessionID {
		return
	}
	w.sessionID = sessionID
	w.state = nil
	w.lastModTime = time.Time{}
	// Immediately try to load
	w.loadLocked()
}

// State returns the current workflow state (may be nil).
func (w *WorkflowWatcher) State() *WorkflowState {
	w.stateMu.RLock()
	defer w.stateMu.RUnlock()
	return w.state
}

// UpdateChannel returns the channel signaled when state changes.
func (w *WorkflowWatcher) UpdateChannel() <-chan struct{} {
	return w.updateCh
}

// Start begins polling the workflow state file.
func (w *WorkflowWatcher) Start() {
	go w.pollLoop()
}

// Close stops the watcher.
func (w *WorkflowWatcher) Close() {
	w.closeOnce.Do(func() {
		close(w.closeCh)
	})
}

func (w *WorkflowWatcher) pollLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.closeCh:
			return
		case <-ticker.C:
			w.check()
		}
	}
}

func (w *WorkflowWatcher) check() {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()

	if w.sessionID == "" {
		return
	}

	path := workflowStatePath(w.profile, w.sessionID)
	if path == "" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		return // file doesn't exist yet
	}

	if !info.ModTime().After(w.lastModTime) {
		return // no change
	}

	w.lastModTime = info.ModTime()
	w.loadLocked()

	// Signal update
	select {
	case w.updateCh <- struct{}{}:
	default:
	}
}

func (w *WorkflowWatcher) loadLocked() {
	if w.sessionID == "" {
		return
	}

	path := workflowStatePath(w.profile, w.sessionID)
	if path == "" {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		w.state = nil // Clear state when file doesn't exist
		return
	}

	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	w.state = &state
}
