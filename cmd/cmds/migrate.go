package cmds

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/migrate"
	// "github.com/springernature/halfpipe/manifest"
	// "github.com/springernature/halfpipe/migrate"
	"github.com/springernature/halfpipe/project"
	"os"
	//	"path"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrates the halfpipe manifest to the latest schema",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkVersion(); err != nil {
			printErr(err)
			os.Exit(1)
		}

		fs := afero.Afero{Fs: afero.NewOsFs()}

		currentDir, err := os.Getwd()
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		projectData, err := project.NewProjectResolver(fs).Parse(currentDir, false)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		controller := createController(projectData, fs, currentDir)
		migrator := migrate.NewMigrator(controller, manifest.Parse, manifest.Render)

		_, migratedYaml, results, migrated, err := migrator.Migrate(projectData.Manifest)
		printErrAndResultAndExitOnError(err, results)

		if migrated {
			fmt.Println("Migrating manifest")
			err := fs.WriteFile(projectData.HalfpipeFilePath, migratedYaml, 0777)
			printErrAndResultAndExitOnError(err, nil)
			fmt.Println("Done")
		} else {
			fmt.Println("Manifest already on latest schema")
		}
	},
}
