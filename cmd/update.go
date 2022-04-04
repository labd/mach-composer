package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/labd/mach-composer-go/updater"
	"github.com/spf13/cobra"
)

var updateFlags struct {
	fileNames     []string
	check         bool
	commit        bool
	commitMessage string
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all (or a given) component.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(generateFlags.fileNames) < 1 {
			matches, err := filepath.Glob("./*.yml")
			if err != nil {
				log.Fatal(err)
			}
			generateFlags.fileNames = matches
			if len(generateFlags.fileNames) < 1 {
				fmt.Println("No .yml files found")
				os.Exit(1)
			}
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := updateFunc(args); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringArrayVarP(&updateFlags.fileNames, "file", "f", nil, "YAML file to update. If not set update all *.yml files.")
	updateCmd.Flags().BoolVarP(&updateFlags.check, "check", "", false, "Only checks for updates, doesnt change files.")
	updateCmd.Flags().BoolVarP(&updateFlags.commit, "commit", "c", false, "Automatically commits the change.")
	updateCmd.Flags().StringVarP(&updateFlags.commitMessage, "commit-message", "m", "", "Use a custom message for the commit.")
}

func updateFunc(args []string) error {
	changes := map[string]string{}

	for _, filename := range updateFlags.fileNames {
		updateSet := updater.UpdateFile(filename)

		if updateSet.HasChanges() {
			changes[filename] = updateSet.ChangeLog()
		}
	}

	if len(changes) < 1 {
		return nil
	}

	// git commit
	if updateFlags.commit {
		filenames := []string{}
		multipleFiles := len(changes) > 1
		commitMessage := updateFlags.commitMessage

		for fn, _ := range changes {
			filenames = append(filenames, fn)
		}

		// Generate commit message if not passed
		if updateFlags.commitMessage == "" {
			var cm strings.Builder
			for fn, msg := range changes {
				if multipleFiles {
					fmt.Fprintf(&cm, "Changes for %s:\n", fn)
					cm.WriteString(msg)
					fmt.Fprintln(&cm, "")
				} else {
					cm.WriteString(msg)
				}
			}
			commitMessage = cm.String()
		}

		ctx := context.Background()
		updater.Commit(ctx, filenames, commitMessage)
	}
	return nil
}
