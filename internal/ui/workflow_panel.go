package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// WorkflowPanel renders workflow graph and artifact registry.
type WorkflowPanel struct {
	width  int
	height int
	state  *WorkflowState
}

// NewWorkflowPanel creates a new workflow panel.
func NewWorkflowPanel() *WorkflowPanel {
	return &WorkflowPanel{}
}

// SetSize sets panel dimensions.
func (p *WorkflowPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetState updates the workflow state to render.
func (p *WorkflowPanel) SetState(state *WorkflowState) {
	p.state = state
}

// View renders the workflow panel content (without the title bar).
func (p *WorkflowPanel) View() string {
	if p.state == nil {
		return p.renderEmpty()
	}

	var b strings.Builder

	// Workflow graph section
	graphLines := p.renderWorkflowGraph()
	b.WriteString(graphLines)

	// Separator between sections
	sepStyle := lipgloss.NewStyle().Foreground(ColorBorder)
	b.WriteString("\n")
	maxSep := p.width
	if maxSep > 0 {
		b.WriteString(sepStyle.Render(strings.Repeat("─", maxSep)))
	}
	b.WriteString("\n")

	// Artifact registry section
	artifactLines := p.renderArtifacts()
	b.WriteString(artifactLines)

	return b.String()
}

func (p *WorkflowPanel) renderEmpty() string {
	dimStyle := lipgloss.NewStyle().Foreground(ColorTextDim).Italic(true)
	msg := dimStyle.Render("No workflow state")
	hint := dimStyle.Render("Configure hooks to enable")

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(msg)
	b.WriteString("\n\n")
	b.WriteString(hint)
	return b.String()
}

func (p *WorkflowPanel) renderWorkflowGraph() string {
	var b strings.Builder

	// Section header
	headerStyle := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	b.WriteString(headerStyle.Render("WORKFLOW"))

	// Status badge
	if p.state.Workflow.Status != "" {
		badge := p.statusBadge(p.state.Workflow.Status)
		b.WriteString("  ")
		b.WriteString(badge)
	}
	b.WriteString("\n")

	if len(p.state.Workflow.Stages) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(ColorTextDim)
		b.WriteString(dimStyle.Render("  No stages defined"))
		return b.String()
	}

	current := p.state.Workflow.Current
	currentIdx := -1
	for i, s := range p.state.Workflow.Stages {
		if s == current {
			currentIdx = i
			break
		}
	}

	for i, stage := range p.state.Workflow.Stages {
		var icon string
		var style lipgloss.Style

		switch {
		case i < currentIdx:
			// Completed stage
			icon = "✓"
			style = lipgloss.NewStyle().Foreground(ColorGreen)
		case i == currentIdx:
			// Current stage
			icon = "●"
			style = lipgloss.NewStyle().Foreground(ColorYellow).Bold(true)
		default:
			// Future stage
			icon = "○"
			style = lipgloss.NewStyle().Foreground(ColorTextDim)
		}

		// Connector line (except for first stage)
		if i > 0 {
			connColor := ColorTextDim
			if i <= currentIdx {
				connColor = ColorGreen
			}
			connStyle := lipgloss.NewStyle().Foreground(connColor)
			b.WriteString(connStyle.Render("  │"))
			b.WriteString("\n")
		}

		// Stage line
		line := fmt.Sprintf("  %s %s", icon, stage)
		b.WriteString(style.Render(line))

		// Current stage indicator
		if i == currentIdx {
			arrowStyle := lipgloss.NewStyle().Foreground(ColorYellow)
			b.WriteString(arrowStyle.Render(" ←"))
		}

		b.WriteString("\n")
	}

	return b.String()
}

func (p *WorkflowPanel) renderArtifacts() string {
	var b strings.Builder

	// Section header with count
	headerStyle := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	countStyle := lipgloss.NewStyle().Foreground(ColorTextDim)

	b.WriteString(headerStyle.Render("ARTIFACTS"))
	if len(p.state.Artifacts) > 0 {
		b.WriteString("  ")
		b.WriteString(countStyle.Render(fmt.Sprintf("%d", len(p.state.Artifacts))))
	}
	b.WriteString("\n")

	if len(p.state.Artifacts) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(ColorTextDim)
		b.WriteString(dimStyle.Render("  None yet"))
		return b.String()
	}

	for _, a := range p.state.Artifacts {
		icon := p.artifactIcon(a.Type)
		titleStyle := lipgloss.NewStyle().Foreground(ColorText)

		// Truncate title to fit width
		maxTitle := p.width - 8 // icon + padding + status
		title := a.Title
		if maxTitle > 0 && len(title) > maxTitle {
			if maxTitle > 3 {
				title = title[:maxTitle-3] + "..."
			}
		}

		b.WriteString(fmt.Sprintf("  %s ", icon))

		// Wrap title in OSC 8 hyperlink if URL is available
		if a.URL != "" {
			b.WriteString(fmt.Sprintf("\x1b]8;;%s\x07", a.URL))
			b.WriteString(titleStyle.Render(title))
			b.WriteString("\x1b]8;;\x07")
		} else {
			b.WriteString(titleStyle.Render(title))
		}

		// Status/ref badge on the right
		badge := p.artifactBadge(a)
		if badge != "" {
			b.WriteString("  ")
			b.WriteString(badge)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (p *WorkflowPanel) statusBadge(status string) string {
	var color lipgloss.Color
	switch status {
	case "active":
		color = ColorGreen
	case "paused":
		color = ColorYellow
	case "completed":
		color = ColorCyan
	case "failed":
		color = ColorRed
	default:
		color = ColorTextDim
	}
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(status)
}

func (p *WorkflowPanel) artifactIcon(typ string) string {
	switch typ {
	case ArtifactTypePR:
		return "⑂"
	case ArtifactTypeCI:
		return "⚙"
	case ArtifactTypeIssue:
		return "◈"
	case ArtifactTypeCommit:
		return "○"
	case ArtifactTypeURL:
		return "◆"
	default:
		return "·"
	}
}

func (p *WorkflowPanel) artifactBadge(a Artifact) string {
	var parts []string

	if a.Status != "" {
		var color lipgloss.Color
		switch a.Status {
		case "passed", "merged", "closed":
			color = ColorGreen
		case "open", "pending", "running":
			color = ColorYellow
		case "failed":
			color = ColorRed
		default:
			color = ColorTextDim
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(color).Render(a.Status))
	}

	if a.Ref != "" && a.Status == "" {
		// Only show ref if no status (e.g. plain commits)
		style := lipgloss.NewStyle().Foreground(ColorComment)
		short := a.Ref
		if len(short) > 7 {
			short = short[:7]
		}
		parts = append(parts, style.Render(short))
	}

	return strings.Join(parts, " ")
}
