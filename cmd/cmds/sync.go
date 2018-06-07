package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/sync"
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs the halfpipe binary to the latest one",
	Run: func(cmd *cobra.Command, args []string) {
		currentVersion, err := config.GetVersion()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		syncer := sync.NewSyncer(currentVersion, sync.ResolveLatestVersionFromArtifactory)
		err = syncer.Update(os.Stdout)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	},
}
