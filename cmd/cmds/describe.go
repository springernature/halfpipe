package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v2"
	"os"
)

type Pipeline struct {
	Name   string `yaml:"name,omitempty"`
	Filter string `yaml:"filter,omitempty"`
}

type CF_App struct {
	Name         string `yaml:"name,omitempty"`
	Space        string `yaml:"space,omitempty"`
	ManifestPath string `yaml:"manifest_path,omitempty"`
	Filter       string `yaml:"filter,omitempty"`
}

type Katee_App struct {
	Name         string `yaml:"name,omitempty"`
	Namespace    string `yaml:"namespace,omitempty"`
	ManifestPath string `yaml:"manifest_path,omitempty"`
	Filter       string `yaml:"filter,omitempty"`
}
type Output struct {
	Usage string `yaml:"usage,omitempty"`
	Team  string `yaml:"team,omitempty"`

	Pipeline   Pipeline    `yaml:"pipeline,omitempty"`
	CF_Apps    []CF_App    `yaml:"cf_apps,omitempty"`
	Katee_Apps []Katee_App `yaml:"katee_apps,omitempty"`
}

func explainPipeline(resp halfpipe.Response) (o Output) {
	o.Usage = "All filters are jq expressions to run against the inventory file at /tmp/inventory.json. If the inventory doesn't exist download it from https://ee-platform.apps.private.k8s.springernature.io/api/v1/inventory and save it to /tmp/inventory.json then run the filter: cat /tmp/inventory.json | jq <filter>"
	o.Team = resp.Manifest.Team
	if resp.Manifest.Platform == "actions" {
		o.Pipeline = Pipeline{
			Name:   resp.Manifest.Pipeline,
			Filter: fmt.Sprintf(`.resources[] | select(.type == "Github Workflow" and .name == "%s" and .team == "%s" and .metadata.repo == "%s")`, resp.Manifest.Pipeline, resp.Manifest.Team, resp.Manifest.Triggers.GetGitTrigger().GetRepoName()),
		}
	} else {
		o.Pipeline = Pipeline{
			Name:   resp.Manifest.Pipeline,
			Filter: fmt.Sprintf(`.resources[] | select(.slug | test("^concourse_%s_%s_[0-9]+$"))`, resp.Manifest.Team, resp.Manifest.Pipeline),
		}
	}

	for _, task := range resp.Manifest.Tasks.Flatten() {
		switch t := task.(type) {
		case manifest.DeployCF:
			o.CF_Apps = append(o.CF_Apps, CF_App{
				Name:         t.Name,
				Space:        t.Space,
				ManifestPath: t.Manifest,
				Filter:       fmt.Sprintf(`.resources[] | select(.slug | test("^cf_.*_%s_%s$"))`, t.Space, t.CfApplication.Name),
			})
		case manifest.DeployKatee:
			app := Katee_App{
				Name:         t.Name,
				Namespace:    "",
				ManifestPath: t.VelaManifest,
				Filter:       "",
			}

			if t.Namespace != "" {
				app.Namespace = t.Namespace
			} else {
				app.Namespace = fmt.Sprintf("katee-%s", resp.Manifest.Team)
			}

			// katee_v2_katee-monetise-live_subscriptions-schedule
			app.Filter = fmt.Sprintf(`.resources[] | select(.slug | test("^katee_v2_%s_%s$"))`, app.Namespace, t.KateeManifest.Metadata.Name)

			o.Katee_Apps = append(o.Katee_Apps, app)
		}
	}

	return o
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: `Prints a description of the pipeline. Handy for ingestion into LLMs`,
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndController(formatInput(Input), nil)
		response, err := controller.Process(man)
		if err != nil {
			printErr(err)
			os.Exit(1)
		}

		if out, err := yaml.Marshal(explainPipeline(response)); err != nil {
			panic(err)
		} else {
			fmt.Println(string(out))
		}
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
