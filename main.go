package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	// Check if GitHub Copilot CLI is installed
	if err := checkCopilotCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Add all changes to staging
	err := executeCommand("git", "add", ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding files to git: %v\n", err)
		os.Exit(1)
	}

	// Get a summary of changes
	changes, err := executeCommandWithOutput("git", "diff", "--cached", "--name-status")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting git diff: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating commit message with Copilot CLI...")
	prompt := fmt.Sprintf("Generate a concise git commit message following conventional commit format (type(scope): description) for these changes. Use types like feat, fix, docs, style, refactor, test, chore. The changes are: %s", changes)

	var commitMsg string
	// Try using gh copilot suggest
	commitMsg, err = generateCommitMessage(prompt)
	if err != nil {
		fmt.Printf("GitHub Copilot CLI error: %v\n", err)
		// Fallback to a basic message
		changedFiles := extractChangedFiles(changes)
		commitMsg = fmt.Sprintf("chore: changes to %s", strings.Join(changedFiles[:min(len(changedFiles), 5)], ", "))
	}

	// Validate and enforce conventional commit format
	commitMsg = enforceConventionalCommit(commitMsg, changes)

	// Commit with the generated message
	fmt.Printf("Committing with message: %s\n", commitMsg)
	err = executeCommand("git", "commit", "-m", commitMsg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error committing changes: %v\n", err)
		os.Exit(1)
	}

	// Push changes
	err = executeCommand("git", "push")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error pushing changes: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Changes pushed successfully!")
}

// enforceConventionalCommit ensures the message follows conventional commit format
func enforceConventionalCommit(message string, changes string) string {
	// Regular expression for conventional commit format
	conventionalFormat := regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\([a-z0-9-]+\))?: .+`)

	// If message already follows the format, return it
	if conventionalFormat.MatchString(message) {
		return message
	}

	// Otherwise, try to determine the appropriate type from the changes
	commitType := determineCommitType(changes)

	// Extract first sentence to use as description
	description := message
	if idx := strings.Index(message, "."); idx > 0 {
		description = message[:idx]
	}
	description = strings.TrimSpace(description)

	// First letter should be lowercase
	if len(description) > 0 {
		description = strings.ToLower(description[:1]) + description[1:]
	}

	return fmt.Sprintf("%s: %s", commitType, description)
}

// determineCommitType tries to determine an appropriate commit type based on changes
func determineCommitType(changes string) string {
	lowerChanges := strings.ToLower(changes)

	// Default type
	commitType := "chore"

	// Try to determine type based on file patterns and change descriptions
	if strings.Contains(lowerChanges, "test") || strings.Contains(lowerChanges, "_test.go") {
		commitType = "test"
	} else if strings.Contains(lowerChanges, "fix") || strings.Contains(lowerChanges, "bug") {
		commitType = "fix"
	} else if strings.Contains(lowerChanges, "feat") || strings.Contains(lowerChanges, "add") ||
		strings.Contains(lowerChanges, "new") {
		commitType = "feat"
	} else if strings.Contains(lowerChanges, "doc") || strings.Contains(lowerChanges, "readme") {
		commitType = "docs"
	} else if strings.Contains(lowerChanges, "refactor") {
		commitType = "refactor"
	} else if strings.Contains(lowerChanges, "style") || strings.Contains(lowerChanges, "format") {
		commitType = "style"
	}

	return commitType
}

// checkCopilotCLI verifies that the GitHub Copilot CLI is installed
func checkCopilotCLI() error {
	cmd := exec.Command("gh", "copilot", "--version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub Copilot CLI is not installed or not accessible. Please install it first: https://github.com/github/gh-copilot")
	}
	return nil
}

func executeCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeCommandWithOutput(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func generateCommitMessage(prompt string) (string, error) {
	// Create a temporary file to store the prompt
	tempFile, err := os.CreateTemp("", "copilot-prompt-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	// Write the prompt to the temporary file
	if _, err = tempFile.WriteString(prompt); err != nil {
		return "", err
	}
	tempFile.Close()

	// Execute GitHub Copilot CLI
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cat %s | gh copilot suggest", tempFile.Name()))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("gh copilot suggest failed: %v: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func extractChangedFiles(changes string) []string {
	lines := strings.Split(changes, "\n")
	var files []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			files = append(files, parts[1])
		}
	}

	return files
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
