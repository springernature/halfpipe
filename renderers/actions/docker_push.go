package actions

import (
	"fmt"
	"path"
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) dockerPushSteps(task manifest.DockerPush, man manifest.Manifest) Steps {
	// If this is an ECR push, use the ECR-specific flow
	if task.IsECR() {
		return a.dockerPushECRSteps(task, man)
	}

	return a.dockerPushGCRSteps(task, man)
}

func (a *Actions) dockerPushGCRSteps(task manifest.DockerPush, man manifest.Manifest) Steps {
	buildArgs := map[string]string{
		"ARTIFACTORY_PASSWORD": "",
		"ARTIFACTORY_URL":      "",
		"ARTIFACTORY_USERNAME": "",
		"BUILD_VERSION":        "",
		"GIT_REVISION":         "",
		"RUNNING_IN_CI":        "",
	}
	for k, v := range task.Vars {
		buildArgs[k] = v
	}

	push := Step{
		Name: "Build and Push",
		Uses: ExternalActions.DockerPush,
		With: With{
			"image":      task.Image,
			"tags":       "latest\n${{ env.BUILD_VERSION }}\n${{ env.GIT_REVISION }}\n",
			"context":    path.Join(a.workingDir, task.BuildPath),
			"dockerfile": path.Join(a.workingDir, task.DockerfilePath),
			"buildArgs":  MultiLine{buildArgs},
			"secrets":    MultiLine{task.Secrets},
			"platforms":  strings.Join(task.Platforms, ","),
		},
	}

	// useCache will be set on manual "workflow dispatch" trigger.
	// otherwise it will be an empty string and we default it to true
	if task.UseCache {
		push.With["useCache"] = "${{ inputs.useCache == '' || inputs.useCache == 'true' }}"
	}

	if man.FeatureToggles.Ghas() {
		push.With["ghas"] = "true"
		push.With["githubPat"] = "${{ secrets.GITHUB_TOKEN }}"
	}

	return Steps{push, repositoryDispatch(task.Image)}
}

func (a *Actions) dockerPushECRSteps(task manifest.DockerPush, _ manifest.Manifest) Steps {
	steps := Steps{}

	// Hardcoded region for China ECR
	const awsRegion = "cn-northwest-1"

	// Step 1: Install AWS CLI
	awsCLI := Step{
		Name: "Install AWS CLI",
		Uses: ExternalActions.AWSCLI,
	}
	steps = append(steps, awsCLI)

	// Step 2: Configure AWS credentials (using GitHub secrets like Artifactory)
	awsCredentials := Step{
		Name: "Configure AWS credentials",
		Uses: ExternalActions.AWSCredentials,
		With: With{
			"aws-access-key-id":     "${{ secrets.AWS_ACCESS_KEY_ID }}",
			"aws-secret-access-key": "${{ secrets.AWS_SECRET_ACCESS_KEY }}",
			"aws-region":            awsRegion,
		},
	}
	steps = append(steps, awsCredentials)

	// Step 3: Login to Amazon ECR
	ecrLogin := Step{
		Name: "Login to Amazon ECR",
		ID:   "login-ecr",
		Uses: ExternalActions.AWSECRLogin,
	}
	steps = append(steps, ecrLogin)

	// Step 4: Create ECR repository if it doesn't exist
	// Extract repository name from image (everything after the registry host)
	repoName := extractECRRepoName(task.Image)
	createRepo := Step{
		Name: "Create ECR repository (if not exists)",
		Run: fmt.Sprintf(`aws ecr describe-repositories --repository-names %s --region %s 2>/dev/null || \
aws ecr create-repository \
  --repository-name %s \
  --region %s \
  --image-scanning-configuration scanOnPush=true`, repoName, awsRegion, repoName, awsRegion),
	}
	steps = append(steps, createRepo)

	// Step 4: Build and push Docker image
	buildArgs := []string{}
	for k, v := range task.Vars {
		if v != "" {
			buildArgs = append(buildArgs, fmt.Sprintf("--build-arg %s=%s", k, v))
		} else {
			buildArgs = append(buildArgs, fmt.Sprintf("--build-arg %s", k))
		}
	}
	// Add default build args
	for _, arg := range []string{"ARTIFACTORY_PASSWORD", "ARTIFACTORY_URL", "ARTIFACTORY_USERNAME", "BUILD_VERSION", "GIT_REVISION", "RUNNING_IN_CI"} {
		buildArgs = append(buildArgs, fmt.Sprintf("--build-arg %s", arg))
	}

	buildArgsStr := strings.Join(buildArgs, " \\\n  ")
	dockerfilePath := path.Join(a.workingDir, task.DockerfilePath)
	buildContext := path.Join(a.workingDir, task.BuildPath)
	if buildContext == "" || buildContext == "." {
		buildContext = a.workingDir
	}

	buildAndPush := Step{
		Name: "Build, tag, and push Docker image to ECR",
		Env: Env{
			"ECR_REGISTRY": "${{ steps.login-ecr.outputs.registry }}",
			"IMAGE_TAG":    "${{ env.BUILD_VERSION }}",
		},
		Run: fmt.Sprintf(`# Build the Docker image
docker build -t $ECR_REGISTRY/%s:$IMAGE_TAG \
  -t $ECR_REGISTRY/%s:${{ env.GIT_REVISION }} \
  -t $ECR_REGISTRY/%s:latest \
  -f %s \
  %s \
  %s

# Push the image to ECR
docker push $ECR_REGISTRY/%s:$IMAGE_TAG
docker push $ECR_REGISTRY/%s:${{ env.GIT_REVISION }}
docker push $ECR_REGISTRY/%s:latest

echo "✅ Image pushed: $ECR_REGISTRY/%s:$IMAGE_TAG"`,
			repoName, repoName, repoName,
			dockerfilePath, buildArgsStr, buildContext,
			repoName, repoName, repoName, repoName),
	}
	steps = append(steps, buildAndPush)

	// Step 5: Repository dispatch (for triggering downstream workflows)
	steps = append(steps, repositoryDispatch(task.Image))

	return steps
}

func extractECRRepoName(image string) string {
	// ECR format: [ACCOUNT_ID].dkr.ecr.[REGION].amazonaws.com[.cn]/[NAMESPACE]/[IMAGE]
	// We need to extract everything after the registry host
	parts := strings.SplitN(image, "/", 2)
	if len(parts) > 1 {
		// Remove any tag if present
		repoName := strings.Split(parts[1], ":")[0]
		return repoName
	}
	return image
}

func repositoryDispatch(name string) Step {
	return Step{
		Name: "Repository dispatch",
		Uses: ExternalActions.RepositoryDispatch,
		With: With{
			"token":      githubSecrets.RepositoryDispatchToken,
			"event-type": "docker-push:" + name,
		},
	}
}
