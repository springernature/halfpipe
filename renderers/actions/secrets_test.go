package actions

import (
	"fmt"
	"strings"
	"testing"
)

func Test_stringListToSecret(t *testing.T) {
	buildArgs := `ARTIFACTORY_PASSWORD=${{ secrets.EE_ARTIFACTORY_PASSWORD }}
ARTIFACTORY_URL=${{ secrets.EE_ARTIFACTORY_URL }}
ARTIFACTORY_USERNAME=${{ secrets.EE_ARTIFACTORY_USERNAME }}
BUILD_VERSION=2.${{ github.run_number }}.0
GIT_REVISION=${{ github.sha }}
RUNNING_IN_CI=true
SECRET=((a.b))
VAULT_ROLE_ID=${{ secrets.VAULT_ROLE_ID }}
VAULT_SECRET_ID=${{ secrets.VAULT_SECRET_ID }}`
	newBuildArgs := `ARTIFACTORY_PASSWORD=${{ secrets.EE_ARTIFACTORY_PASSWORD }}
ARTIFACTORY_URL=${{ secrets.EE_ARTIFACTORY_URL }}
ARTIFACTORY_USERNAME=${{ secrets.EE_ARTIFACTORY_USERNAME }}
BUILD_VERSION=2.${{ github.run_number }}.0
GIT_REVISION=${{ github.sha }}
RUNNING_IN_CI=true
SECRET=${{ steps.secrets.outputs.springernature_data_blah_a_b }}
VAULT_ROLE_ID=${{ secrets.VAULT_ROLE_ID }}
VAULT_SECRET_ID=${{ secrets.VAULT_SECRET_ID }}`

	got, args := multiLineStringToSecret(strings.Split(buildArgs, "\n"), "blah")

	if len(got) == 0 {
		t.Errorf("Expected a secret but got none")
	}

	if args != newBuildArgs {
		t.Errorf(fmt.Sprintf("got %s but wanted %s", args, newBuildArgs))
	}
	for _, secret := range got {
		if secret.vaultPath != "/springernature/data/blah/a b" {
			t.Errorf(fmt.Sprintf("got: %v, want: %v", secret, "blah"))
		}
	}
}
