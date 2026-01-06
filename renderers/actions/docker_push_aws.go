package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
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
  --encryption-configuration encryptionType=AES256`, task.Repository, task.Repository),
	}

	buildAndPush := Step{
		Name: "Build, tag, and push Docker image",
		Env: Env{
			"ECR_REGISTRY": "${{ steps.login-ecr.outputs.registry }}",
			"IMAGE_TAG":    "${{ github.sha }}",
		},
		Run: fmt.Sprintf(`docker build -t $ECR_REGISTRY/%s:$IMAGE_TAG -f %s %s
docker push $ECR_REGISTRY/%s:$IMAGE_TAG
docker tag $ECR_REGISTRY/%s:$IMAGE_TAG $ECR_REGISTRY/%s:latest
docker push $ECR_REGISTRY/%s:latest`, task.Repository, dockerfilePath, buildPath, task.Repository, task.Repository, task.Repository, task.Repository),
	}

	return Steps{configureAWS, loginECR, installAWSCLI, createECRRepo, buildAndPush}
}
