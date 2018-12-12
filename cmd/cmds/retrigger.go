package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe/retrigger"
	"os"
)

func init() {
	retriggerCmd := &cobra.Command{
		Use:   "retrigger-errored-builds-for-team",
		Short: "Retriggers all the errored builds in a team",
		Long: `Sometimes there are a lot of errored builds, maybe Github or GCP were down, maybe we just did a mistake?

In those cases it would be handy to retrigger all errored builds for a team, as Concourse does not do it automatically.

Note that you must be logged into the team with fly before executing this and that team must == target..
 
Essentially what it does is

builds = getAllBuilds(--team)[0:--count]
for build in builds:
  if build.errored:
    # if there has been any builds after this we should not retrigger
    if build.IsLatestForPipelineJob:
      build.Retrigger()`,
		Run: func(cmd *cobra.Command, args []string) {
			team := cmd.Flag("team").Value.String()
			count := cmd.Flag("count").Value.String()

			builds, err := retrigger.GetBuilds(team, count)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}

			erroredBuilds := builds.GetErrored()
			if len(erroredBuilds) == 0 {
				fmt.Println("No errored builds found")
				os.Exit(0)
			}

			buildsToBeRetriggered := retrigger.Builds{}
			for _, erroredBuild := range erroredBuilds {
				if builds.IsLatest(erroredBuild) {
					buildsToBeRetriggered = append(buildsToBeRetriggered, erroredBuild)
				}
			}

			if len(buildsToBeRetriggered) == 0 {
				fmt.Println("I found the following errored builds")
				for _, erroredBuild := range erroredBuilds {
					fmt.Println(fmt.Sprintf("  * %s [#%s]", erroredBuild, erroredBuild.Name))
				}
				fmt.Println("But there has been newer builds for each of them so I will just ignore them.")
				os.Exit(0)
			}

			for _, buildToBeRetriggered := range buildsToBeRetriggered {
				fmt.Printf("\nGoing to retrigger %s\n", buildToBeRetriggered)
				err := buildToBeRetriggered.Retrigger()
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
				fmt.Println("")
			}

			return
		},

		PreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("team").Value.String() == "" {
				fmt.Println("--team must be set!")
				os.Exit(-1)
			}
		},
	}

	rootCmd.AddCommand(retriggerCmd)
	retriggerCmd.Flags().String("team", "", "team you want to retrigger errored builds for")
	retriggerCmd.Flags().Int("count", 100, "only look at the last count builds, defaults to 100")
}
