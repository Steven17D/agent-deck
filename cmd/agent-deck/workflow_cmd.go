package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

func handleWorkflow(profile string, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: agent-deck workflow <command> [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  check    Check if workflow state was updated (for use as Claude Code hook)")
		fmt.Fprintln(os.Stderr, "  path     Print the workflow state file path for a session")
		os.Exit(2)
	}

	switch args[0] {
	case "check":
		handleWorkflowCheck(profile, args[1:])
	case "path":
		handleWorkflowPath(profile, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown workflow command: %s\n", args[0])
		os.Exit(2)
	}
}

// handleWorkflowCheck is called by Claude Code hooks (PostToolUse, Stop).
// It checks if the workflow state file was updated since the last check.
// If not, it prints a reminder to stderr so Claude sees it.
func handleWorkflowCheck(profile string, args []string) {
	var sessionID string
	var onStop bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--session", "-s":
			if i+1 < len(args) {
				sessionID = args[i+1]
				i++
			}
		case "--on-stop":
			onStop = true
		}
	}

	if sessionID == "" {
		// Try environment variable
		sessionID = os.Getenv("AGENTDECK_INSTANCE_ID")
	}

	if sessionID == "" {
		// No session ID — silently skip (not running inside agent-deck)
		return
	}

	profileDir, err := session.GetProfileDir(profile)
	if err != nil {
		os.Exit(2)
	}

	stateFile := filepath.Join(profileDir, "workflow", sessionID+".json")
	lastCheckFile := filepath.Join(profileDir, "workflow", sessionID+".lastcheck")

	// Ensure workflow directory exists
	_ = os.MkdirAll(filepath.Dir(stateFile), 0o755)

	// Get state file mtime
	stateInfo, err := os.Stat(stateFile)
	if err != nil {
		// State file doesn't exist — block until Claude creates it
		fmt.Fprintf(os.Stderr, "Run /workflow to create your workflow state file at %s before stopping.\n", stateFile)
		os.Exit(2)
	}

	// Get last check time
	lastCheckTime := time.Time{}
	if checkInfo, err := os.Stat(lastCheckFile); err == nil {
		lastCheckTime = checkInfo.ModTime()
	}

	// If state file hasn't been modified since last check, remind Claude
	if !stateInfo.ModTime().After(lastCheckTime) {
		if onStop {
			fmt.Fprintf(os.Stderr, "Run /workflow to update your workflow state at %s before stopping.\n", stateFile)
			os.Exit(2) // Block stop until Claude updates the file
		} else {
			fmt.Fprintf(os.Stderr, "Reminder: Update your workflow state at %s. Set current stage and register any new artifacts (commits, PRs, test results).\n", stateFile)
		}
		return
	}

	// Only update last check timestamp on success
	_ = os.MkdirAll(filepath.Dir(lastCheckFile), 0o755)
	_ = os.WriteFile(lastCheckFile, []byte(time.Now().Format(time.RFC3339)), 0o644)
}

// handleWorkflowPath prints the workflow state file path for a session.
func handleWorkflowPath(profile string, args []string) {
	var sessionID string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--session", "-s":
			if i+1 < len(args) {
				sessionID = args[i+1]
				i++
			}
		}
	}

	if sessionID == "" {
		sessionID = os.Getenv("AGENTDECK_INSTANCE_ID")
	}

	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "Error: --session or AGENTDECK_INSTANCE_ID required")
		os.Exit(2)
	}

	profileDir, err := session.GetProfileDir(profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	fmt.Println(filepath.Join(profileDir, "workflow", sessionID+".json"))
}
