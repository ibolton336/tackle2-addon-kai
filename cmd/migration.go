package main

import (
	"fmt"
	"os"
	"os/exec"
)

// RunMigration runs goose with the migration skill against the source directory.
// It reads LLM config from environment (AWS Bedrock via GOOSE_PROVIDER etc.)
// and runs goose non-interactively with the skill as context.
func RunMigration(sourceDir string, skillPath string) (err error) {
	if _, err = exec.LookPath("goose"); err != nil {
		return fmt.Errorf("goose binary not found in PATH: %w", err)
	}

	skillContent, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("failed to read skill: %w", err)
	}

	prompt := fmt.Sprintf(`You are performing a Java EE to Quarkus migration.

Read and follow the migration guide below carefully.

%s

Now migrate the Java EE application in the current directory to Quarkus 3.x.
Work through the codebase systematically. Commit your changes as you go.
When complete, summarize what was changed.`, string(skillContent))

	cmd := exec.Command("goose", "run", "--text", prompt)
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	if os.Getenv("GOOSE_PROVIDER") == "" {
		cmd.Env = append(cmd.Env, "GOOSE_PROVIDER=aws_bedrock")
	}
	if os.Getenv("GOOSE_MODEL") == "" {
		cmd.Env = append(cmd.Env, "GOOSE_MODEL=us.anthropic.claude-sonnet-4-5-20250929-v1:0")
	}
	if os.Getenv("BEDROCK_ENABLE_CACHING") == "" {
		cmd.Env = append(cmd.Env, "BEDROCK_ENABLE_CACHING=false")
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("goose migration failed: %w", err)
	}
	return nil
}

// PushMigrationBranch pushes the migration changes to a new branch.
func PushMigrationBranch(sourceDir string, branchName string) (err error) {
	commands := [][]string{
		{"git", "config", "user.email", "kai-addon@konveyor.io"},
		{"git", "config", "user.name", "Kai Migration Addon"},
		{"git", "checkout", "-b", branchName},
		{"git", "add", "-A"},
		{"git", "commit", "-m", "chore: kai automated migration to quarkus"},
		{"git", "push", "origin", branchName},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = sourceDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("git command %v failed: %w", args, err)
		}
	}
	return nil
}
