package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"os"
)

func init() {
	rootCmd.AddCommand(internalRepresentation)
}

type nullRenderer struct{}

func (r nullRenderer) Render(manifest manifest.Manifest) (string, error) {
	return "", nil
}

var internalRepresentation = &cobra.Command{
	Use:   "internal-representation",
	Short: `Prints the internal representation of the manifest`,
	Run: func(cmd *cobra.Command, args []string) {

		man, controller := getManifestAndController(formatInput(Input))

		defaultedAndMappedManifest, _ := controller.DefaultAndMap(man)

		updatedManifest, err := yaml.Marshal(defaultedAndMappedManifest)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		fmt.Println(string(updatedManifest))
	},
}
