package ui

import (
	"strings"
	"testing"
)

func TestWorkflowPanelRender(t *testing.T) {
	InitTheme("dark")

	p := NewWorkflowPanel()
	p.SetSize(30, 20)

	// Test empty state
	p.SetState(nil)
	out := p.View()
	if !strings.Contains(out, "No workflow state") {
		t.Errorf("expected empty state message, got: %s", out)
	}

	// Test with state
	state := &WorkflowState{
		Version:   1,
		SessionID: "test-123",
		Workflow: Workflow{
			Stages:  []string{"Planning", "Investigation", "Implementation", "Testing", "PR"},
			Current: "Implementation",
			Status:  "active",
		},
		Artifacts: []Artifact{
			{Type: ArtifactTypeCommit, Title: "feat: add structs", Ref: "a1b2c3d"},
			{Type: ArtifactTypeCI, Title: "Jenkins #184", URL: "https://jenkins.example.com/184", Status: "passed"},
			{Type: ArtifactTypePR, Title: "PR #42", URL: "https://github.com/org/repo/pull/42", Ref: "#42", Status: "open"},
			{Type: ArtifactTypeIssue, Title: "ANJ-1905", URL: "https://jira.example.com/ANJ-1905", Ref: "ANJ-1905", Status: "open"},
			{Type: ArtifactTypeURL, Title: "Design Doc", URL: "https://docs.example.com/design"},
		},
	}

	p.SetState(state)
	out = p.View()

	// Check workflow section
	if !strings.Contains(out, "WORKFLOW") {
		t.Error("missing WORKFLOW header")
	}
	if !strings.Contains(out, "active") {
		t.Error("missing status badge")
	}
	if !strings.Contains(out, "Planning") {
		t.Error("missing Planning stage")
	}
	if !strings.Contains(out, "Implementation") {
		t.Error("missing Implementation stage")
	}
	// Check completed stages have checkmark
	if !strings.Contains(out, "✓") {
		t.Error("missing checkmark for completed stages")
	}
	// Check current stage marker
	if !strings.Contains(out, "●") {
		t.Error("missing current stage marker")
	}
	// Check future stages
	if !strings.Contains(out, "○") {
		t.Error("missing future stage marker")
	}

	// Check artifacts section
	if !strings.Contains(out, "ARTIFACTS") {
		t.Error("missing ARTIFACTS header")
	}
	if !strings.Contains(out, "feat: add structs") {
		t.Error("missing commit artifact")
	}
	if !strings.Contains(out, "a1b2c3d") {
		t.Error("missing commit ref")
	}
	if !strings.Contains(out, "Jenkins #184") {
		t.Error("missing CI artifact")
	}
	if !strings.Contains(out, "passed") {
		t.Error("missing CI status")
	}
	if !strings.Contains(out, "PR #42") {
		t.Error("missing PR artifact")
	}
	if !strings.Contains(out, "open") {
		t.Error("missing PR status")
	}
	if !strings.Contains(out, "ANJ-1905") {
		t.Error("missing issue artifact")
	}
	if !strings.Contains(out, "Design Doc") {
		t.Error("missing URL artifact")
	}

	t.Logf("Rendered output:\n%s", out)
}
