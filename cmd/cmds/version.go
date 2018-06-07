package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/config"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := config.GetVersion()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(version.String())
	},
}
