package session

import (
	"os"
	"path/filepath"
)

const workflowSkillContent = `# Workflow State Update

Maintain your workflow state file for agent-deck's workflow panel.

## Instructions

### 1. Find the state file path

` + "```bash" + `
agent-deck workflow path
` + "```" + `

This prints the path to your workflow state file (e.g., ` + "`~/.agent-deck/profiles/default/workflow/<session-id>.json`" + `).

If the command returns nothing, ` + "`AGENTDECK_INSTANCE_ID`" + ` is not set — you are not running inside agent-deck. Skip this workflow.

### 2. Read or create the file

If the file exists, read it. If not, create it.

The schema:

` + "```json" + `
{
  "version": 1,
  "session_id": "<AGENTDECK_INSTANCE_ID>",
  "claude_session_id": "<your Claude session ID from the status bar>",
  "workflow": {
    "stages": ["<stage1>", "<stage2>", "..."],
    "current": "<current stage>",
    "status": "active"
  },
  "artifacts": []
}
` + "```" + `

**Stages** — Define stages that match what you're actually doing. Common patterns:
- Feature work: ` + "`" + `["Planning", "Implementation", "Testing", "PR", "Review", "Merge"]` + "`" + `
- Bug fix: ` + "`" + `["Triage", "Investigation", "Fix", "Test", "PR"]` + "`" + `
- Research: ` + "`" + `["Exploration", "Analysis", "Report"]` + "`" + `
- Simple task: ` + "`" + `["Working", "Done"]` + "`" + `

Choose stages that reflect the actual task, not a generic template.

### 3. Update the file

Set ` + "`current`" + ` to whatever stage you're in right now. Update ` + "`status`" + `:
- ` + "`\"active\"`" + ` — working on it
- ` + "`\"paused\"`" + ` — stopping, will resume later
- ` + "`\"completed\"`" + ` — done
- ` + "`\"failed\"`" + ` — blocked or abandoned

Add artifacts as you produce them. Each artifact must have a well-defined type:

| Type | Description | Key fields |
|------|-------------|------------|
| ` + "`pr`" + ` | Pull request (GitHub, Bitbucket, GitLab) | url, ref (#number), status (open/merged/closed) |
| ` + "`ci`" + ` | CI/CD run (Jenkins, GitHub Actions, etc.) | url, status (running/passed/failed) |
| ` + "`issue`" + ` | Issue tracker (Jira, GitHub Issues, Linear) | url, ref (issue key), status (open/closed) |
| ` + "`commit`" + ` | Git commit | ref (SHA) |
| ` + "`url`" + ` | Generic URL (docs, dashboards, etc.) | url |

` + "```json" + `
{"type": "commit", "title": "fix: null pointer", "ref": "a1b2c3d"}
{"type": "pr", "title": "PR #42", "url": "https://github.com/org/repo/pull/42", "ref": "#42", "status": "open"}
{"type": "ci", "title": "Jenkins #184", "url": "https://jenkins.example.com/184", "status": "passed"}
{"type": "issue", "title": "ANJ-1905", "url": "https://jira.example.com/ANJ-1905", "ref": "ANJ-1905", "status": "open"}
{"type": "url", "title": "Design Doc", "url": "https://docs.example.com/design"}
` + "```" + `

### 4. Write the file

Write the updated JSON to the path from step 1. The agent-deck TUI will pick up changes within 1 second.

### When to run this

- At the **start** of a session — create the file with initial stages
- When you **change stage** — update ` + "`current`" + `
- When you **produce an artifact** — add to artifacts array
- When **stopping** — set status to ` + "`\"paused\"`" + ` or ` + "`\"completed\"`" + `
- When the **Stop hook blocks you** — it means you forgot to update. Run this, then the hook will pass.
`

// installWorkflowSkill installs the /workflow skill to ~/.claude/commands/.
func installWorkflowSkill(configDir string) {
	// configDir is typically ~/.claude — install to commands/ under it
	commandsDir := filepath.Join(configDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return
	}

	skillPath := filepath.Join(commandsDir, "workflow.md")

	// Don't overwrite if user has customized it
	if _, err := os.Stat(skillPath); err == nil {
		return
	}

	_ = os.WriteFile(skillPath, []byte(workflowSkillContent), 0644)
}
