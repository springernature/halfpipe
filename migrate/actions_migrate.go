package migrate

import (
	"fmt"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
	"os"
	"path"
	"path/filepath"
)

func ActionsMigrationHelper(man manifest.Manifest, response halfpipe.Response) (err error) {
	fmt.Println("To migrate a Concourse pipeline to Actions you must do the following steps")

	fmt.Println("1. Read the docs over at https://ee.public.springernature.app/rel-eng/github-actions/overview/")

	fmt.Println("2. Pause the Concourse pipeline")
	fmt.Printf("\t $ fly -t %s pause-pipeline -p %s\n", man.Team, man.Pipeline)

	fmt.Println("3. Add the necessary Vault tokens to the Github repo's secrets. If you have a mono repo you only need to do this once")
	fmt.Printf("\t $ vault read springernature/%s/team-ro-app-role\n", man.Team)
	fmt.Printf("\t # Write the secrets as VAULT_ROLE_ID and VAULT_SECRET_ID in https://github.com/springernature/%s/settings/secrets/actions\n", response.Project.RootName)

	fmt.Println("4. Update the halfpipe file with the new top level key, 'platform: actions'")

	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	relPath, err := filepath.Rel(pwd, response.Project.GitRootPath)
	if err != nil {
		return
	}

	fmt.Println("5. Render, commit and push the pipeline")
	fmt.Println("\t $ halfpipe")
	fmt.Printf("\t $ git add %s\n", path.Join(relPath, ".github", "workflows", man.PipelineName()+".yml"))
	fmt.Printf("\t $ git commit -m \"Added workflow for %s\"\n", man.PipelineName())
	fmt.Println("\t $ git push")

	fmt.Println("6. Watch the pipeline in all its glory https://github.com/springernature/halfpipe-examples/actions?query=workflow%3A" + man.PipelineName())

	return
}
