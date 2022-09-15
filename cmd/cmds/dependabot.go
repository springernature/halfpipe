package cmds

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/dependabot"
	"log"
	"os"
)

func init() {
	var depth int
	var verbose bool
	var skipFolders []string

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	logrus.SetOutput(os.Stderr)

	dependabotCmd := &cobra.Command{
		Use:   "dependabot",
		Short: "Creates a dependabot config",
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}
			fs := afero.NewOsFs()
			pwd, err := os.Getwd()
			if err != nil {
				log.Panic(err)
			}

			dependabot.New(
				dependabot.DependabotConfig{Depth: depth, Verbose: verbose, SkipFolders: skipFolders},
				dependabot.NewWalker(afero.Afero{Fs: afero.NewBasePathFs(fs, pwd)}),
			).Resolve()
		},
	}

	dependabotCmd.Flags().IntVar(&depth, "depth", 3, "Max depth to scan.")
	dependabotCmd.Flags().BoolVar(&verbose, "verbose", false, "Print verbose information")
	dependabotCmd.Flags().StringSliceVar(&skipFolders, "skip-folder", []string{}, "Skipped folders relative from root")

	rootCmd.AddCommand(dependabotCmd)
}
