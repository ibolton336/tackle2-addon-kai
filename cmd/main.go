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
	Repository      repository.SCM
	Source          string
	MigrationTarget string `json:"migrationTarget"`
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
		migrationTarget := d.MigrationTarget
		if migrationTarget == "" {
			migrationTarget = "java-ee-to-quarkus"
		}
		skillPath := path.Join(Dir, "skills", migrationTarget, "SKILL.md")

		// Run migration with goose
		addon.Activity("Running migration with goose.")
		err = RunMigration(SourceDir, skillPath)
		if err != nil {
			return
		}

		// Push migration branch
		addon.Activity("Pushing migration branch.")
		branchName := fmt.Sprintf("kai/migrate-to-quarkus-%d", addon.Task.ID)
		err = PushMigrationBranch(SourceDir, branchName)
		if err != nil {
			return
		}

		addon.Activity("Migration complete.")
		return
	})
}
