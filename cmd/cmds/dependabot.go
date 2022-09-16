package cmds

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/dependabot"
	"gopkg.in/yaml.v2"
	"os"
)

func init() {
	var depth int
	var verbose bool
	var skipFolders []string
	var skipEcosystem []string

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
				logrus.Panic(err)
			}

			c, err := dependabot.New(
				dependabot.NewWalker(afero.Afero{Fs: afero.NewBasePathFs(fs, pwd)}, depth, skipFolders),
				dependabot.NewFilter(skipEcosystem),
				dependabot.NewRender(),
			).Resolve()
			if err != nil {
				logrus.Panic(err)
			}

			out, err := yaml.Marshal(c)
			if err != nil {
				logrus.Panic(err)
			}
			fmt.Println(string(out))
		},
	}

	dependabotCmd.Flags().IntVar(&depth, "depth", 3, "Max depth to scan.")
	dependabotCmd.Flags().BoolVar(&verbose, "verbose", false, "Print verbose information")
	dependabotCmd.Flags().StringSliceVar(&skipFolders, "skip-folder", []string{}, "Skipped folders relative from root")
	dependabotCmd.Flags().StringSliceVar(&skipEcosystem, "skip-ecosystem", []string{}, "Skipped ecosystems. Find them at https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file#package-ecosystem")

	rootCmd.AddCommand(dependabotCmd)
}
