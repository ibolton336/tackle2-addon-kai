package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// RunMigration runs goose with the migration skill against the source directory.
// It reads LLM config from environment (AWS Bedrock via GOOSE_PROVIDER etc.)
// and runs goose non-interactively with the skill as context.
//
// Architecture: two-layer skill system
//   - Orchestrator (/addon/skills/SKILL.md): imperative workflow instructions
//   - Transformation ref (/addon/skills/<target>/SKILL.md): specific mappings
//   - Pallet-synced skills (.goose/skills/): runtime-fetched overrides
func RunMigration(sourceDir string, skillPath string, d *Data) (err error) {
	if _, err = exec.LookPath("goose"); err != nil {
		return fmt.Errorf("goose binary not found in PATH: %w", err)
	}

	// Load orchestrator skill (generic imperative workflow)
	orchestratorPath := path.Join(path.Dir(skillPath), "..", "SKILL.md")
	orchestratorContent, orchErr := os.ReadFile(orchestratorPath)
	if orchErr != nil {
		// Fall back: use skill path directly if no orchestrator exists
		orchestratorContent = nil
	}

	// Load transformation reference (migration-type-specific)
	// Priority: pallet-synced > baked-in
	palletSkillPath := path.Join(sourceDir, ".goose", "skills", d.MigrationTarget, "SKILL.md")
	skillContent, readErr := os.ReadFile(palletSkillPath)
	if readErr != nil {
		// Fall back to baked-in skill
		skillContent, err = os.ReadFile(skillPath)
		if err != nil {
			return fmt.Errorf("failed to read skill: %w", err)
		}
	}

	// Build task configuration section
	var skipMsg string
	if len(d.SkipPackages) > 0 {
		skipMsg = strings.Join(d.SkipPackages, ", ")
	} else {
		skipMsg = "none"
	}

	dryRunMsg := ""
	if d.DryRun {
		dryRunMsg = "\n- THIS IS A DRY RUN: describe all planned changes in detail but do NOT write or modify any files"
	}

	taskConfig := fmt.Sprintf(`
## Task Configuration (these override skill defaults)
- Migration target: %s
- Messaging provider: %s
- Java target version: %s
- Packages to skip (do not modify): %s%s
`, d.MigrationTarget, d.MessagingProvider, d.JavaTarget, skipMsg, dryRunMsg)

	// Assemble the full prompt: orchestrator + transformation ref + task config
	var prompt string
	if orchestratorContent != nil {
		prompt = fmt.Sprintf(`%s

---

# Transformation Reference (%s)

%s
%s
Migrate the application in the current directory NOW. Start with Phase 1.`,
			string(orchestratorContent), d.MigrationTarget, string(skillContent), taskConfig)
	} else {
		// Legacy fallback: skill-only prompt
		prompt = fmt.Sprintf(`You are performing a code migration.

Read and follow the migration guide below carefully.

%s
%s
Now migrate the application in the current directory.
Work through the codebase systematically. Transform source files in place.
Do NOT write documentation or guide files. Only modify source code.
When complete, summarize what was changed.`, string(skillContent), taskConfig)
	}

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

// RunPalletSync runs pallet sync in the source directory to pull skills
// from configured sources into .goose/skills/.
func RunPalletSync(sourceDir string) error {
	palletPath, err := exec.LookPath("pallet")
	if err != nil {
		return fmt.Errorf("pallet binary not found: %w", err)
	}

	// Check if pallet.yaml exists in the source dir
	palletYaml := path.Join(sourceDir, "pallet.yaml")
	if _, err := os.Stat(palletYaml); os.IsNotExist(err) {
		// Check env var for pallet config content
		palletContent := os.Getenv("PALLET_YAML")
		if palletContent == "" {
			return fmt.Errorf("no pallet.yaml found and PALLET_YAML not set")
		}
		// Write it
		if err := os.WriteFile(palletYaml, []byte(palletContent), 0644); err != nil {
			return fmt.Errorf("failed to write pallet.yaml: %w", err)
		}
	}

	cmd := exec.Command(palletPath, "sync", ".")
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pallet sync failed: %w", err)
	}
	return nil
}
