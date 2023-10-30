package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
)

func init() {
	rootCmd.AddCommand(internalRepresentation)
}

var internalRepresentation = &cobra.Command{
	Use:   "internal-representation",
	Short: `Prints the internal representation of the manifest`,
	Run: func(cmd *cobra.Command, args []string) {

		man, controller := getManifestAndController(formatInput(Input), nil)

		defaultedAndMappedManifest, _ := controller.DefaultAndMap(man)

		updatedManifest, err := yaml.Marshal(defaultedAndMappedManifest)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		fmt.Println(string(updatedManifest))
	},
}
