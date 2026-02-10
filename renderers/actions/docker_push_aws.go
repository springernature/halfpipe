package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared/secrets"
)

func (a *Actions) dockerPushAWSSteps(task manifest.DockerPushAWS, man manifest.Manifest) Steps {
	dockerfilePath := task.DockerfilePath
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	buildPath := task.BuildPath
	if buildPath == "" {
		buildPath = "."
	}

	configureAWS := Step{
		Name: "Configure AWS credentials",
		Uses: ExternalActions.AWSConfigureCredentials,
		With: With{
			"aws-access-key-id":     task.AccessKeyID,
			"aws-secret-access-key": task.SecretAccessKey,
			"aws-region":            task.Region,
		},
	}

	loginECR := Step{
		Name: "Login to Amazon ECR",
		ID:   "login-ecr",
		Uses: ExternalActions.AWSECRLogin,
	}

	installAWSCLI := Step{
		Name: "Install AWS CLI",
		Run: `curl -sL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip -q awscliv2.zip
sudo ./aws/install --update
rm -rf awscliv2.zip aws/`,
	}

	createECRRepo := Step{
		Name: "Create ECR repository (if not exists)",
		Run: fmt.Sprintf(`aws ecr describe-repositories --repository-names %s 2>/dev/null || \
aws ecr create-repository \
  --repository-name %s \
  --image-scanning-configuration scanOnPush=true \
  --encryption-configuration encryptionType=AES256`, task.Image, task.Image),
	}

	buildAndPush := a.dockerPushAWSBuildStep(task, dockerfilePath, buildPath)

	return Steps{configureAWS, loginECR, installAWSCLI, createECRRepo, buildAndPush}
}

func (a *Actions) dockerPushAWSBuildStep(task manifest.DockerPushAWS, dockerfilePath, buildPath string) Step {
	buildArgs := map[string]string{
		"ARTIFACTORY_PASSWORD": "",
		"ARTIFACTORY_URL":      "",
		"ARTIFACTORY_USERNAME": "",
		"BUILD_VERSION":        "",
		"GIT_REVISION":         "",
		"RUNNING_IN_CI":        "",
		"CI":                   "",
	}

	env := Env{
		"ECR_REGISTRY": "${{ steps.login-ecr.outputs.registry }}",
		"IMAGE_TAG":    "${{ github.sha }}",
	}

	for k, v := range task.Vars {
		if secrets.IsSecret(v) {
			env[k] = v
			buildArgs[k] = ""
		} else {
			buildArgs[k] = v
		}
	}

	dockerBuildCmd := a.buildDockerBuildCommand(task.Image, dockerfilePath, buildPath, buildArgs, task.Secrets)

	for k, v := range task.Secrets {
		env[k] = v
	}

	return Step{
		Name: "Build, tag, and push Docker image",
		Env:  env,
		Run: fmt.Sprintf(`%s
docker push $ECR_REGISTRY/%s:$IMAGE_TAG
docker tag $ECR_REGISTRY/%s:$IMAGE_TAG $ECR_REGISTRY/%s:latest
docker push $ECR_REGISTRY/%s:latest`, dockerBuildCmd, task.Image, task.Image, task.Image, task.Image),
	}
}

func (a *Actions) buildDockerBuildCommand(image, dockerfilePath, buildPath string, buildArgs map[string]string, secrets manifest.Vars) string {
	var parts []string
	parts = append(parts, "docker build")

	var argKeys []string
	for k := range buildArgs {
		argKeys = append(argKeys, k)
	}
	sort.Strings(argKeys)

	for _, k := range argKeys {
		v := buildArgs[k]
		if v == "" {
			parts = append(parts, fmt.Sprintf("--build-arg %s", k))
		} else {
			parts = append(parts, fmt.Sprintf("--build-arg %s=%s", k, v))
		}
	}

	var secretKeys []string
	for k := range secrets {
		secretKeys = append(secretKeys, k)
	}
	sort.Strings(secretKeys)

	for _, k := range secretKeys {
		parts = append(parts, fmt.Sprintf("--secret id=%s,env=%s", k, k))
	}

	parts = append(parts, fmt.Sprintf("-t $ECR_REGISTRY/%s:$IMAGE_TAG", image))
	parts = append(parts, fmt.Sprintf("-f %s", dockerfilePath))
	parts = append(parts, buildPath)

	return strings.Join(parts, " \\\n  ")
}
