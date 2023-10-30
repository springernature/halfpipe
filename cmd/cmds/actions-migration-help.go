package cmds

import (
	"github.com/spf13/cobra"
	"os"
	"text/template"
)

func init() {
	rootCmd.AddCommand(actionsMigrationHelp)
}

var actionsMigrationHelp = &cobra.Command{
	Use:   "actions-migration-help",
	Short: "Prints out the steps needed to migrate from Concourse to Actions",
	Run: func(cmd *cobra.Command, args []string) {
		man, controller := getManifestAndController(formatInput(Input), nil)
		response, _ := controller.Process(man)

		tpl, _ := template.New("").Parse(`
To migrate a Concourse pipeline to Actions you must do the following steps

1. Read the docs over at https://ee.public.springernature.app/rel-eng/github-actions/overview/

2. Pause the Concourse pipeline
   $ fly -t {{.Team}} pause-pipeline -p {{.Pipeline}}

3. Add the necessary Vault tokens to the Github repo's secrets. If you have a mono repo you only need to do this once
   $ vault kv get springernature/{{.Team}}/team-ro-app-role
   # Write the secrets as VAULT_ROLE_ID and VAULT_SECRET_ID in https://github.com/springernature/{{.Project}}/settings/secrets/actions

4. Update the halfpipe file with the new top level key, 'platform: actions'

5. Render, commit and push the pipeline
   $ halfpipe
   $ git add .
   $ git commit -m "Added github actions workflow for {{.Pipeline}}"
   $ git push

6. View the workflow on GitHub
   $ halfpipe url

7. Delete the old Concourse pipeline
   $ fly -t {{.Team}} destroy-pipeline -p {{.Pipeline}}
`)

		data := struct {
			Team, Pipeline, Project string
		}{
			Team:     man.Team,
			Pipeline: man.PipelineName(),
			Project:  response.Project.RootName,
		}

		_ = tpl.Execute(os.Stdout, data)

	},
}
