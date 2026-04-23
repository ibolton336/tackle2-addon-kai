package main

import (
	"fmt"
	"os"
	"path"

	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-addon/ssh"
	hub "github.com/konveyor/tackle2-hub/addon"
)

var (
	addon     = hub.Addon
	Dir       = ""
	SourceDir = ""
	Source    = "Kai"
)

type Data struct {
	Repository        repository.SCM
	Source            string
	MigrationTarget   string   `json:"migrationTarget"`   // skill name, default "java-ee-to-quarkus"
	MessagingProvider string   `json:"messagingProvider"` // "in-memory", "rabbitmq", "kafka" — default "in-memory"
	JavaTarget        string   `json:"javaTarget"`        // "17", "21" — default "17"
	SkipPackages      []string `json:"skipPackages"`      // packages to leave untouched
	BranchName        string   `json:"branchName"`        // output branch — default "kai/migrate-to-quarkus"
	DryRun            bool     `json:"dryRun"`            // plan only, do not write files
}

func init() {
	Dir, _ = os.Getwd()
	SourceDir = path.Join(Dir, "source")
}

func main() {
	addon.Run(func() (err error) {
		d := &Data{}
		err = addon.DataWith(d)
		if err != nil {
			return
		}
		if d.Source == "" {
			d.Source = Source
		}
		if d.MigrationTarget == "" {
			d.MigrationTarget = "java-ee-to-quarkus"
		}
		if d.JavaTarget == "" {
			d.JavaTarget = "17"
		}
		if d.MessagingProvider == "" {
			d.MessagingProvider = "in-memory"
		}
		if d.BranchName == "" {
			d.BranchName = "kai/migrate-to-quarkus"
		}

		//
		// Fetch application.
		addon.Activity("Fetching application.")
		application, err := addon.Task.Application()
		if err != nil {
			return
		}

		// Pallet sync skills and log hashes
		// use archetype to get appropriate skills?
		// arch, err := addon.Task.Archetype()

		// Fetch assessment to get questions for usage in mig plan?

		// SSH agent
		agent := ssh.Agent{}
		err = agent.Start()
		if err != nil {
			return
		}
		// Fetch source repo
		err = FetchRepository(application)
		if err != nil {
			return
		}

		// Load migration skill
		addon.Activity("Loading migration skill.")
		skillPath := path.Join(Dir, "skills", d.MigrationTarget, "SKILL.md")

		// Run migration with goose
		addon.Activity("Running migration with goose.")
		err = RunMigration(SourceDir, skillPath, d)
		if err != nil {
			return
		}

		// Push migration branch (skip if dry run)
		if d.DryRun {
			addon.Activity("Dry run complete.")
		} else {
			addon.Activity("Pushing migration branch.")
			err = PushMigrationBranch(SourceDir, d.BranchName)
			if err != nil {
				return
			}
			addon.Activity("Migration complete.")
		}
		return
	})
}
