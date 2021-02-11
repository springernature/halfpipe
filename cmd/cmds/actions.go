package cmds

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(actionsCmd)
}

var actionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "**NO LONGER SUPPORTED** Replaced with 'platform: actions' in the halfpipe manifest.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, color.FgRed.Sprint("[ERROR] The 'halfpipe actions' command is no longer supported. Set 'platform: actions' in the halfpipe manifest and run 'halfpipe'."))
		os.Exit(1)
	},
}
