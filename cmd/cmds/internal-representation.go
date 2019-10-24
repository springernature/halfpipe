package cmds

import (
	"fmt"
	"github.com/simonjohansson/yaml"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(internalRepresentation)
}

var internalRepresentation = &cobra.Command{
	Use:   "internal-representation",
	Short: ``,
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndCreateController()

		defaultedAndMappedManifest := controller.DefaultAndMap(man)

		updatedManifest, err := yaml.Marshal(defaultedAndMappedManifest)
		printErrAndResultAndExitOnError(err, nil)

		fmt.Println(string(updatedManifest))
	},
}
