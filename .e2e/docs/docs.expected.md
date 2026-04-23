# Halfpipe Manifest Reference

## Contents

- [Top-Level Fields](#top-level-fields)
- [Triggers](#triggers)
  - [docker](#docker-trigger)
  - [git](#git-trigger)
  - [pipeline](#pipeline-trigger)
  - [timer](#timer-trigger)
- [Tasks](#tasks)
  - [buildpack](#buildpack)
  - [consumer-integration-test](#consumer-integration-test)
  - [copy-container-image](#copy-container-image)
  - [deploy-cf](#deploy-cf)
  - [deploy-katee](#deploy-katee)
  - [deploy-ml-modules](#deploy-ml-modules)
  - [deploy-ml-zip](#deploy-ml-zip)
  - [docker-compose](#docker-compose)
  - [docker-push](#docker-push)
  - [parallel](#parallel)
  - [run](#run)
  - [sequence](#sequence)
- [Supporting Types](#supporting-types)
  - [notifications](#notifications)
  - [notification channel](#notification-channel)
  - [vars](#vars)
  - [docker](#docker)
  - [github_environment](#github_environment)
  - [feature_toggles](#feature_toggles)

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `team` | string | required | The platform team that owns this pipeline. |
| `pipeline` | string | required | The name of the pipeline. |
| `platform` | `concourse`, `actions` | optional | The CI platform to target. Defaults to concourse. |
| `triggers` | [Trigger](#triggers)[] | optional | The triggers that cause this pipeline to run. Defaults to git. |
| `tasks` | [Task](#tasks)[] | required | The tasks that make up this pipeline. |
| `notifications` | [notifications](#notifications) | optional | Default notifications for all tasks. |
| `feature_toggles` | [feature_toggles](#feature_toggles) | optional | Optional feature toggles |
| `teams_webhook` | string | optional | A Microsoft Teams webhook URL for pipeline-level notifications. |
| `slack_channel` | string | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |
| `slack_failure_message` | string | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |
| `slack_success_message` | string | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |


## Triggers

Triggers cause the pipeline to run. Specified under the `triggers` key.

### `docker` (trigger)

docker trigger runs the pipeline when a docker image has been updated.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | required | Docker image to watch for updates. |
| `username` | string | optional | Username for private Docker registries. |
| `password` | string | optional | Password for private Docker registries. |

**Example:**

```yaml
# Trigger when a docker image is updated
- type: docker
  image: "eu.gcr.io/halfpipe-io/halfpipe-example-docker"
```

### `git` (trigger)

git trigger defines which git repo halfpipe will operate on. By convention
there is always a git trigger as default. To disable it, set manual_trigger
to true.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `uri` | string | optional | Git repository URI. Defaults to the URI resolved from .git/config. |
| `private_key` | string | optional | SSH private key for cloning the repository. Defaults to ((github.private_key)). |
| `watched_paths` | string[] | optional | Only trigger when changes occur in these paths (globs supported). |
| `ignored_paths` | string[] | optional | Do not trigger when changes occur only in these paths (globs supported). |
| `git_crypt_key` | string | optional | Base64-encoded git-crypt key to unlock an encrypted repository. |
| `branch` | string | optional | Branch to track. Required when running halfpipe on a non-default branch. |
| `shallow` | boolean | optional | Shallow clone the repository (--depth 1). Defaults to false in Concourse and true in GitHub Actions. |
| `manual_trigger` | boolean | optional | Disable automatic triggering on commits. |

**Examples:**

```yaml
# Override the default uri and private key
- type: git
  uri: git@github.com:org/repo.git
  private_key: ((repo-name.private-key))
```

```yaml
# Only trigger when there are changes in src/main,
# unlock the encrypted repo, and shallow clone.
- type: git
  uri: git@github.com:organisation/repo-name.git
  private_key: ((repo-name.private-key))
  git_crypt_key: ((git-crypt-keys.repo-name))
  watched_paths:
    - src/main
  shallow: true
```

```yaml
# Disable automatic git triggering, use a cron timer instead
- type: git
  manual_trigger: true
- type: timer
  cron: "0 8 * * *"
```

### `pipeline` (trigger)

pipeline trigger runs the pipeline when another pipeline job has completed.
Note that you cannot trigger on pipelines from another team.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `concourse_url` | string | optional | Concourse URL. Defaults to the current Concourse instance. |
| `username` | string | optional | Concourse username. |
| `password` | string | optional | Concourse password. |
| `team` | string | optional | Team that owns the pipeline to trigger from. Must be the same team. |
| `pipeline` | string | required | Name of the pipeline to trigger from. |
| `job` | string | required | Job name within the pipeline to trigger from. |
| `status` | string | optional | Job status to trigger on. Allowed values: succeeded, failed, errored, aborted. Defaults to succeeded. |

**Example:**

```yaml
# Trigger when another pipeline job fails
- type: pipeline
  pipeline: my-cool-pipeline
  job: Deploy to SNPaaS
  status: failed
```

### `timer` (trigger)

timer trigger runs the pipeline on a schedule. The cron expression must be
valid; remember to specify times in UTC. See [crontab.guru] for help
writing cron expressions.

[crontab.guru]: https://crontab.guru/

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cron` | string | required | Cron expression for the timer schedule. Times must be in UTC. |

**Example:**

```yaml
# Run every 10 minutes on weekdays
- type: timer
  cron: "*/10 * * * 1-5"
```


## Tasks

Tasks define the steps in your pipeline. Specified under the `tasks` key.

### `buildpack`

buildpack generates a container image using Cloud Native Buildpacks and
publishes it to the Halfpipe registry. The task uses [Paketo Buildpacks]
which is an implementation of the Cloud Native Buildpacks specification.

[Paketo Buildpacks]: https://paketo.io

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `image` | string | required | Docker image name to build and push. Format: eu.gcr.io/halfpipe-io/<team>/<image-name>. |
| `buildpacks` | string[] | required | Buildpack identifiers to use for building the image e.g. paketo-buildpacks/java. |
| `builder` | string | optional | Paketo builder to use. Defaults to paketobuildpacks/builder-jammy-buildpackless-base. |
| `path` | string | optional | Path to the application source code to build. Defaults to current directory. |
| `restore_artifacts` | boolean | optional | Restore artifacts saved by previous tasks. |
| `vars` | [vars](#vars) | optional | Environment variables passed to the pack build command. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: buildpack
  image: eu.gcr.io/halfpipe-io/my-team/my-app
  buildpacks:
    - paketo-buildpacks/java
```

```yaml
# More complex with custom builder, multiple buildpacks and vars
- type: buildpack
  image: eu.gcr.io/halfpipe-io/my-team/my-app
  builder: paketobuildpacks/builder-jammy-buildpackless-full
  buildpacks:
    - paketo-buildpacks/java
    - paketo-buildpacks/nodejs
  path: target/my-app.jar
  vars:
    ENV1: 1234
    ENV2: ((secret.something))
    ENV3: '{"a": "b", "c": "d"}'
    API_KEY: ((api.key))
```

### `consumer-integration-test`

consumer-integration-test is designed to run in a provider's pipeline. The
task allows for a test script to be run. The script is passed two environment
variables automatically: DEPENDENCY_NAME (set by provider_name) and
<DEPENDENCY_NAME>_DEPLOYED_HOST (set by provider_host).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `consumer` | string | required | GitHub repository name of the consumer, with optional sub-directory for monorepos e.g. repo-name or monorepo/dir. |
| `consumer_host` | string | required | Address of the consumer application in the same environment as the provider. |
| `script` | string | required | Consumer test script to execute. |
| `git_clone_options` | string | optional | Custom options for git clone of the consumer repository e.g. --depth 100. |
| `provider_host` | string | optional | Address of the provider application to test. Defaults to the candidate route in pre_promote. |
| `provider_name` | string | optional | Name of the provider app, exposed as DEPENDENCY_NAME. Defaults to the pipeline name. |
| `docker_compose_file` | string | optional | Path to the consumer docker-compose file. Defaults to docker-compose.yml. |
| `docker_compose_service` | string | optional | Service name in the consumer docker-compose. Defaults to code. |
| `vars` | [vars](#vars) | optional | Environment variables available to the docker-compose service. |
| `use_covenant` | boolean | optional | Enable Covenant contract testing support. |
| `save_artifacts` | string[] | optional | Paths to files or directories to save for use in subsequent tasks. |
| `save_artifacts_on_failure` | string[] | optional | Paths to save when the task fails, useful for test reports. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# In pre_promote stage (TEST_ROUTE is injected automatically)
tasks:
  - type: deploy-cf
    space: dev
    pre_promote:
      - type: consumer-integration-test
        name: example consumer tests
        consumer: consumer-repo/optional-sub-directory
        consumer_host: consumer-a.dev.private.springernature.io
        script: ci/run-external-and-cdcs-dev
        docker_compose_service: app
```

```yaml
# Standalone with explicit provider_host
- type: consumer-integration-test
  name: example consumer tests
  consumer: consumer-repo/optional-sub-directory
  consumer_host: consumer-a.dev.private.springernature.io
  provider_host: provider-a.dev.private.springernature.io
  script: ci/run-external-and-cdcs-dev
```

### `copy-container-image`

copy-container-image copies an image from the halfpipe registry
(eu.gcr.io/halfpipe-io/) to another registry. Currently only AWS ECR is
supported as the target. Normally this would be used after a docker-push
or buildpack task.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `source` | string | required | Full source image URL in the halfpipe registry, with or without tag. |
| `target` | string | required | Target ECR image URL or bare ECR registry URL. |
| `aws_access_key_id` | string | optional | AWS access key ID for the target ECR registry. Defaults to shared credentials from Vault. |
| `aws_secret_access_key` | string | optional | AWS secret access key for the target ECR registry. Defaults to shared credentials from Vault. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Copy image using BUILD_VERSION tag (default)
- type: buildpack
  image: eu.gcr.io/halfpipe-io/my-team/my-app
  buildpacks:
    - paketo-buildpacks/java

- type: copy-container-image
  source: eu.gcr.io/halfpipe-io/my-team/my-app
  target: 1234567890.dkr.ecr.cn-northwest-1.amazonaws.com.cn
```

```yaml
# Copy to a custom target path and tag
- type: docker-push
  image: eu.gcr.io/halfpipe-io/my-team/my-app

- type: copy-container-image
  source: eu.gcr.io/halfpipe-io/my-team/my-app
  target: 1234567890.dkr.ecr.cn-northwest-1.amazonaws.com.cn/another-team/another-image:1.0.0
```

### `deploy-cf`

deploy-cf deploys an app to Cloud Foundry.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `space` | string | required | Cloud Foundry space to deploy to. |
| `api` | string | optional | Cloud Foundry API endpoint. Defaults to ((cloudfoundry.api-snpaas)). |
| `org` | string | optional | Cloud Foundry organisation. Defaults to the value of team. |
| `username` | string | optional | Cloud Foundry username. Defaults to ((cloudfoundry.username)). |
| `password` | string | optional | Cloud Foundry password. Defaults to ((cloudfoundry.password)). |
| `manifest` | string | optional | Path to the Cloud Foundry app manifest, relative to the halfpipe manifest. Defaults to manifest.yml. |
| `test_domain` | string | optional | Domain used when pushing the app as a candidate. Derived from the API by default. |
| `vars` | [vars](#vars) | optional | Environment variables injected into the CF app environment. |
| `deploy_artifact` | string | optional | Path to a file or directory saved by a previous task to deploy to CF. |
| `pre_promote` | [Task](#tasks)[] | optional | Tasks to run after the candidate is deployed but before it is promoted to live. TEST_ROUTE is injected. |
| `pre_start` | string[] | optional | CF CLI commands to run immediately before the candidate app is started. |
| `rolling` | boolean | optional | Use rolling deployment instead of blue-green. |
| `stop_candidate_on_failure` | boolean | optional | Stop the candidate app if deployment fails. |
| `cli_version` | string | optional | CF CLI version to use. Allowed values: cf7, cf8. Defaults to cf7. |
| `docker_tag` | string | optional | Docker image tag to deploy. Required when deploying a Docker image: version or gitref. |
| `sso_route` | string | optional | Route to configure with SSO. |
| `github_environment` | [github_environment](#github_environment) | optional | GitHub environment to associate with this deployment. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: deploy-cf
  space: test
```

```yaml
# More complex with vars, pre_start and pre_promote
- type: deploy-cf
  name: deploy to live
  api: ((cloudfoundry.api-snpaas))
  org: engineering-enablement
  space: live
  manifest: ci/manifest.yml
  vars:
    API_ENDPOINT: https://api.com
    SKIP_SSL_CHECK: true
    APP_SECRET: ((myapp.app_secret_name))
  deploy_artifact: target/distribution/artifact.zip
  pre_start:
    - cf add-network-policy myapp-CANDIDATE --destination-app myapp-CANDIDATE --protocol tcp --port 7600
    - cf events myapp-CANDIDATE
  pre_promote:
    - type: run
      name: run-smoke-tests
      script: ./smoke.sh
      docker:
        image: alpine
```

```yaml
# Deploy a docker image from CF manifest
- type: deploy-cf
  name: deploy docker img
  api: ((cloudfoundry.api-snpaas))
  space: live
  docker_tag: version
```

### `deploy-katee`

deploy-katee deploys an application to Katee.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `vars` | [vars](#vars) | optional | Environment variables available to the vela manifest. |
| `vela_manifest` | string | optional | Path to the vela manifest. Defaults to vela.yaml. |
| `tag` | string | optional | ⚠️ Deprecated: no longer used - safe to delete. |
| `environment` | string | optional | ⚠️ Deprecated: no longer used - safe to delete. |
| `namespace` | string | optional | Vela namespace to deploy to. Defaults to katee-<team>. |
| `deployment_check_timeout` | integer | optional | ⚠️ Deprecated: use max_checks and check_interval instead. |
| `check_interval` | integer | optional | Seconds between each deployment status check. Defaults to 2. |
| `max_checks` | integer | optional | Maximum number of status checks before the deployment is considered failed. Defaults to 60. |
| `github_environment` | [github_environment](#github_environment) | optional | GitHub environment to associate with this deployment. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: deploy-katee
```

```yaml
# More complex with namespace, vars and custom checks
- type: deploy-katee
  namespace: katee-springerlink-prod
  vela_manifest: ./config/vela-manifest-prod.yml
  check_interval: 5
  max_checks: 12
  vars:
    ENV1: 1234
    ENV2: ((secret.something))
    ENV3: '{"a": "b", "c": "d"}'
    API_KEY: ((api.key))
```

### `deploy-ml-modules`

deploy-ml-modules deploys a version of the shared ml modules library from
artifactory.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `ml_modules_version` | string | required | Version of the ml-modules artifact in Artifactory. |
| `targets` | string[] | required | MarkLogic instances to deploy to. |
| `app_name` | string | optional | App name in MarkLogic. Defaults to the pipeline name. |
| `app_version` | string | optional | App version in MarkLogic. Defaults to the git revision. Cannot be set with use_build_version. |
| `use_build_version` | boolean | optional | Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version. |
| `username` | string | optional | Username to connect to MarkLogic. Defaults to the shared vault secret. |
| `password` | string | optional | Password to connect to MarkLogic. Defaults to the shared vault secret. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: deploy-ml-modules
  ml_modules_version: "2.1428"
  targets:
    - marklogic.host
```

```yaml
# Complete with custom app name, version and multiple targets
- type: deploy-ml-modules
  name: deploy xquery - dev
  ml_modules_version: "2.1428"
  app_name: example-app
  app_version: v1
  targets:
    - marklogic.dev.host
    - marklogic.qa.host
    - marklogic.live.host
```

### `deploy-ml-zip`

deploy-ml-zip deploys local XQuery files to MarkLogic using ml-deploy.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `deploy_zip` | string | required | Path to the zip file containing XQuery source files, relative to the manifest. |
| `targets` | string[] | required | MarkLogic instances to deploy to. |
| `app_name` | string | optional | App name in MarkLogic. Defaults to the pipeline name. |
| `app_version` | string | optional | App version in MarkLogic. Defaults to the git revision. Cannot be set with use_build_version. |
| `use_build_version` | boolean | optional | Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version. |
| `username` | string | optional | Username to connect to MarkLogic. Defaults to the shared vault secret. |
| `password` | string | optional | Password to connect to MarkLogic. Defaults to the shared vault secret. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: deploy-ml-zip
  deploy_zip: target/xquery.zip
  targets:
    - marklogic.host
```

```yaml
# Complete with custom app name, version and multiple targets
- type: deploy-ml-zip
  name: deploy xquery - dev
  deploy_zip: target/xquery.zip
  app_name: example-app
  app_version: v1
  targets:
    - marklogic.dev.host
    - marklogic.qa.host
    - marklogic.live.host
```

### `docker-compose`

docker-compose executes docker-compose based on a docker-compose.yml file.
This file must be present in the same directory as the halfpipe manifest.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `command` | string | optional | Command to run against the service. If omitted the default service command is used. |
| `vars` | [vars](#vars) | optional | Environment variables available to docker-compose. |
| `service` | string | optional | Name of the docker-compose service to run. Defaults to app. |
| `compose_file` | string | optional | Space-separated list of docker-compose files |
| `save_artifacts` | string[] | optional | Paths to files or directories to save for use in subsequent tasks. |
| `restore_artifacts` | boolean | optional | Restore artifacts saved by previous tasks. |
| `save_artifacts_on_failure` | string[] | optional | Paths to save when the task fails, useful for test reports. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: docker-compose
```

```yaml
# With name and vars
- type: docker-compose
  name: run tests
  vars:
    TEST_API: https://test-api.com
    MY_SECRET: ((my-app.my-secret-in-vault))
```

### `docker-push`

docker-push builds a Docker image and pushes it to a docker registry. The
image will be tagged with the latest tag, the gitref and pipeline version
by default.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `image` | string | required | Docker image to build and push. Recommended format: eu.gcr.io/halfpipe-io/<team>/<image-name>. |
| `username` | string | optional | Username for the target Docker registry. |
| `password` | string | optional | Password for the target Docker registry. |
| `ignore_vulnerabilities` | boolean | optional | Do not fail the build if critical vulnerabilities are found during image scanning. |
| `scan_timeout` | integer | optional | Number of minutes a Trivy vulnerability scan is allowed to run before timing out. |
| `vars` | [vars](#vars) | optional | Docker build-time variables (ARGs). Do not use for secrets - values are visible in docker history. |
| `secrets` | [vars](#vars) | optional | Docker build-time secrets, mounted securely during build. |
| `restore_artifacts` | boolean | optional | Restore artifacts saved by previous tasks. |
| `dockerfile_path` | string | optional | Path to the Dockerfile, relative to the manifest. Defaults to Dockerfile. |
| `build_path` | string | optional | Path to the folder to use as the Docker build context, relative to the manifest. |
| `tag` | string | optional | ⚠️ Deprecated: no longer used - safe to delete. |
| `platforms` | string[] | optional | Target platforms to build for, e.g. linux/amd64, linux/arm64. Defaults to linux/amd64. |
| `use_cache` | boolean | optional | Enable layer caching to speed up builds by reusing layers from previous builds. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Using the Halfpipe Private Registry (no credentials needed)
- type: docker-push
  image: eu.gcr.io/halfpipe-io/team/image-name
```

```yaml
# Using official Docker Hub
- type: docker-push
  name: push to docker hub
  username: username
  password: ((my.password))
  image: username/image-name
```

```yaml
# Using relative paths for build dir and Dockerfile
- type: docker-push
  name: push GCR
  build_path: buildFolder
  dockerfile_path: ../ops/dockerfiles/Dockerfile
  image: eu.gcr.io/halfpipe-io/team/image-name
```

### `parallel`

parallel enables running tasks in parallel. All tasks start simultaneously;
the group succeeds when all complete.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tasks` | [Task](#tasks)[] | required | Tasks to run in parallel. All tasks start simultaneously; the group succeeds when all complete. |

**Example:**

```yaml
# Build, then deploy to dev and QA in parallel
tasks:
  - type: run
    name: build
    script: build.sh
    docker:
      image: golang
  - type: parallel
    tasks:
      - type: deploy-cf
        name: deploy to dev
        space: dev
      - type: deploy-cf
        name: deploy to QA
        space: qa
  - type: parallel
    tasks:
      - type: deploy-cf
        name: deploy live staging
        space: live-staging
      - type: deploy-cf
        name: deploy live
        space: live
```

### `run`

run is the most generic piece of work you can do. It represents a job in a
pipeline where a script will be run in a docker container. If the script
returns a non-zero exit code the task will be considered failed and any
subsequent tasks will not run.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Optional display name. |
| `script` | string | required | Path to the script to execute, relative to the manifest. Prefix with \ to run a system command e.g. \make. |
| `docker` | [docker](#docker) | required | Docker configuration for the task to run in. |
| `privileged` | boolean | optional | Run the task as root. Not recommended but sometimes necessary e.g. for docker-in-docker. |
| `vars` | [vars](#vars) | optional | Environment variables available to the script. |
| `save_artifacts` | string[] | optional | Paths to files or directories to save for use in subsequent tasks. |
| `restore_artifacts` | boolean | optional | Restore artifacts saved by previous tasks. |
| `save_artifacts_on_failure` | string[] | optional | Paths to save when the task fails, useful for test reports. |
| `manual_trigger` | boolean | optional | Task must be triggered manually (Concourse only). |
| `retries` | integer | optional | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | optional | ⚠️ Deprecated: use notifications instead. |
| `notifications` | [notifications](#notifications) | optional | Notification channels for this task. |
| `timeout` | string | optional | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | optional | Number of build logs to retain. Defaults to 20 (Concourse only). |

**Examples:**

```yaml
# Minimal
- type: run
  script: test.sh
  docker:
    image: golang
```

```yaml
# More complex with vars, artifacts and private registry
- type: run
  name: run tests
  script: test.sh
  docker:
    image: golang
    username: user1
    password: very-secret
  vars:
    TEST_API: https://test-api.com
    MY_SECRET: ((myapp.my-secret-in-vault))
  save_artifacts:
    - target/distribution/artifact.zip
  save_artifacts_on_failure:
    - testReports
```

```yaml
# Restore artifacts from a previous task
- type: run
  script: build.sh
  docker:
    image: eu.gcr.io/halfpipe-io/your-private-image
  restore_artifacts: true
```

```yaml
# Run a system command instead of a script
- type: run
  name: Run uptime from the container
  script: \uptime
  docker:
    image: eu.gcr.io/halfpipe-io/your-private-image
```

### `sequence`

sequence enables running tasks in sequence within a parallel group. It can
only be used inside a parallel task.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tasks` | [Task](#tasks)[] | required | Tasks to run in sequence within a parallel group. Can only be used inside a parallel task. |

**Example:**

```yaml
# Sequence inside parallel:
#
#        +----b1----b2----\
#        |                 \
#  a-----|----c1----c2----c3----e
#        |                 /
#        +----d-----------/
#
tasks:
  - type: run
    name: a
  - type: parallel
    tasks:
      - type: sequence
        tasks:
          - type: run
            name: b1
          - type: run
            name: b2
      - type: sequence
        tasks:
          - type: run
            name: c1
          - type: run
            name: c2
          - type: run
            name: c3
      - type: run
        name: d
  - type: run
    name: e
```


## Supporting Types

### `notifications`

notifications configure which channels to notify on task success or failure.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `success` | [notification channel](#notification-channel)[] | optional | Notification channels to notify on task success. |
| `failure` | [notification channel](#notification-channel)[] | optional | Notification channels to notify on task failure. |
| `on_success` | string[] | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |
| `on_success_message` | string | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |
| `on_failure` | string[] | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |
| `on_failure_message` | string | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |

### `notification channel`

notification channel defines where to send a notification.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `teams` | string | optional | Microsoft Teams channel webhook URL. |
| `message` | string | optional | Optional message to include in the notification. |
| `slack` | string | optional | ⚠️ Deprecated: Slack notifications are no longer supported. |

**Example:**

```yaml
# Notify a one channel success and two channels on failure
notifications:
  success:
    - teams: https://platform-api.ee.springernature.io/api/v1/message?team=my-team
  failure:
    - teams: https://platform-api.ee.springernature.io/api/v1/message?channel=channel.id
      message: Deployment failed - please investigate.
    - teams: ((my-app.teams-webhook))
```

### `vars`

Key-value pairs of environment variables (values are coerced to strings)

**Example:**

```yaml
vars:
  PORT: 8080
  DEBUG: true
  LOG_LEVEL: "error"
  VAULT_SECRET: ((my-app.my-secret-in-vault))
```

### `docker`

Docker image configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | required | Path of docker image in registry including tag. |
| `username` | string | optional | Username for private Docker registries. |
| `password` | string | optional | Password for private Docker registries. |

**Examples:**

```yaml
# Halfpipe or public registry
docker:
  image: eu.gcr.io/halfpipe-io/team/image:tag
```

```yaml
# Private registry
docker:
  image: eu.gcr.io/halfpipe-io/my-team/my-image
  username: ((registry.username))
  password: ((registry.password))
```

### `github_environment`

GitHub environment to associate with this deployment.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | optional | Name of the GitHub environment. |
| `url` | string | optional | URL associated with the GitHub environment. |

**Example:**

```yaml
# Associate a GitHub environment with a deployment
github_environment:
  name: production
  url: https://my-app.example.com
```

### `feature_toggles`

Enable optional pipeline behaviours.

| Toggle | Description |
|--------|-------------|
| `update-pipeline` | Inserts a job that keeps the pipeline/workflow in sync with the halfpipe manifest. Sets BUILD_VERSION. |
| `update-pipeline-and-tag` | Like update-pipeline, but also tags the git repo with `<PIPELINE_NAME>/v<BUILD_VERSION>`. |
| `github-statuses` | Updates GitHub commit statuses from Concourse job results (Actions does this by default). |
| `ghas` | Enables GitHub Advanced Security scanning on docker-push tasks. |

**Example:**

```yaml
# Enable multiple features
feature_toggles:
  - update-pipeline
  - ghas
```
