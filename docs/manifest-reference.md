# Halfpipe Manifest Reference

## Contents

- [Top-Level Fields](#top-level-fields)
- [Triggers](#triggers)
  - [`docker`](#docker-trigger)
  - [`git`](#git-trigger)
  - [`pipeline`](#pipeline-trigger)
  - [`timer`](#timer-trigger)
- [Tasks](#tasks)
  - [`buildpack`](#buildpack)
  - [`consumer-integration-test`](#consumer-integration-test)
  - [`copy-container-image`](#copy-container-image)
  - [`deploy-cf`](#deploy-cf)
  - [`deploy-katee`](#deploy-katee)
  - [`deploy-ml-modules`](#deploy-ml-modules)
  - [`deploy-ml-zip`](#deploy-ml-zip)
  - [`docker-compose`](#docker-compose)
  - [`docker-push`](#docker-push)
  - [`parallel`](#parallel)
  - [`run`](#run)
  - [`sequence`](#sequence)
- [Supporting Types](#supporting-types)
  - [Notifications](#notifications)
  - [NotificationChannel](#notificationchannel)
  - [Vars](#vars)
  - [Docker](#docker)
  - [GitHubEnvironment](#githubenvironment)
  - [Feature Toggles](#feature-toggles)

## Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `team` | string | The platform team that owns this pipeline. |
| `pipeline` | string | The name of the pipeline. |
| `platform` | `concourse`, `actions` | The CI platform to target. Defaults to concourse. |
| `triggers` | Trigger[] | The triggers that cause this pipeline to run. |
| `tasks` | Task[] | The tasks that make up this pipeline. |
| `notifications` | [Notifications](#notifications) | Default notifications for all tasks. |
| `feature_toggles` | string[] | Optional feature toggles |
| `teams_webhook` | string | A Microsoft Teams webhook URL for pipeline-level notifications. |
| `slack_channel` | string | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `slack_failure_message` | string | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `slack_success_message` | string | **Deprecated.** Deprecated: Slack notifications are no longer supported. |


## Triggers

Triggers cause the pipeline to run. Specified under the `triggers` key.

### `docker` (trigger)

| Field | Type | Description |
|-------|------|-------------|
| `image` | string | Docker image to watch for updates. |
| `username` | string | Username for private Docker registries. |
| `password` | string | Password for private Docker registries. |

### `git` (trigger)

| Field | Type | Description |
|-------|------|-------------|
| `uri` | string | Git repository URI. Defaults to the URI resolved from .git/config. |
| `private_key` | string | SSH private key for cloning the repository. Defaults to ((github.private_key)). |
| `watched_paths` | string[] | Only trigger when changes occur in these paths (globs supported). |
| `ignored_paths` | string[] | Do not trigger when changes occur only in these paths (globs supported). |
| `git_crypt_key` | string | Base64-encoded git-crypt key to unlock an encrypted repository. |
| `branch` | string | Branch to track. Required when running halfpipe on a non-default branch. |
| `shallow` | boolean | Shallow clone the repository (--depth 1). Defaults to false in Concourse and true in GitHub Actions. |
| `manual_trigger` | boolean | Disable automatic triggering on commits. |

### `pipeline` (trigger)

| Field | Type | Description |
|-------|------|-------------|
| `concourse_url` | string | Concourse URL. Defaults to the current Concourse instance. |
| `username` | string | Concourse username. |
| `password` | string | Concourse password. |
| `team` | string | Team that owns the pipeline to trigger from. Must be the same team. |
| `pipeline` | string | Name of the pipeline to trigger from. |
| `job` | string | Job name within the pipeline to trigger from. |
| `status` | string | Job status to trigger on. Allowed values: succeeded, failed, errored, aborted. Defaults to succeeded. |

### `timer` (trigger)

| Field | Type | Description |
|-------|------|-------------|
| `cron` | string | Cron expression for the timer schedule. Times must be in UTC. |


## Tasks

Tasks define the steps in your pipeline. Specified under the `tasks` key.

### `buildpack`

| Field | Type | Description |
|-------|------|-------------|
| `builder` | string | Paketo builder to use. Defaults to paketobuildpacks/builder-jammy-buildpackless-base. |
| `buildpacks` | string[] | Buildpack identifiers to use for building the image e.g. paketo-buildpacks/java. |
| `path` | string | Path to the application source code to build. Defaults to current directory. |
| `image` | string | Docker image name to build and push. Format: eu.gcr.io/halfpipe-io/<team>/<image-name>. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `name` | string | Optional display name. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `manual_trigger` | boolean | Task must be triggered manually (Concourse only). |
| `restore_artifacts` | boolean | Restore artifacts saved by previous tasks. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `vars` | [Vars](#vars) | Environment variables passed to the pack build command. |

### `consumer-integration-test`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `consumer` | string | GitHub repository name of the consumer, with optional sub-directory for monorepos e.g. repo-name or monorepo/dir. |
| `consumer_host` | string | Address of the consumer application in the same environment as the provider. |
| `git_clone_options` | string | Custom options for git clone of the consumer repository e.g. --depth 100. |
| `provider_host` | string | Address of the provider application to test. Defaults to the candidate route in pre_promote. |
| `provider_name` | string | Name of the provider app, exposed as DEPENDENCY_NAME. Defaults to the pipeline name. |
| `script` | string | Consumer test script to execute. |
| `docker_compose_file` | string | Path to the consumer docker-compose file. Defaults to docker-compose.yml. |
| `docker_compose_service` | string | Service name in the consumer docker-compose. Defaults to code. |
| `vars` | [Vars](#vars) | Environment variables available to the docker-compose service. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |
| `use_covenant` | boolean | Enable Covenant contract testing support. |
| `save_artifacts` | string[] | Paths to files or directories to save for use in subsequent tasks. |
| `save_artifacts_on_failure` | string[] | Paths to save when the task fails, useful for test reports. |

### `copy-container-image`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |
| `aws_access_key_id` | string | AWS access key ID for the target ECR registry. Defaults to shared credentials from Vault. |
| `aws_secret_access_key` | string | AWS secret access key for the target ECR registry. Defaults to shared credentials from Vault. |
| `source` | string | Full source image URL in the halfpipe registry, with or without tag. |
| `target` | string | Target ECR image URL or bare ECR registry URL. |

### `deploy-cf`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `api` | string | Cloud Foundry API endpoint. Defaults to ((cloudfoundry.api-snpaas)). |
| `space` | string | Cloud Foundry space to deploy to. |
| `org` | string | Cloud Foundry organisation. Defaults to the value of team. |
| `username` | string | Cloud Foundry username. Defaults to ((cloudfoundry.username)). |
| `password` | string | Cloud Foundry password. Defaults to ((cloudfoundry.password)). |
| `manifest` | string | Path to the Cloud Foundry app manifest, relative to the halfpipe manifest. Defaults to manifest.yml. |
| `test_domain` | string | Domain used when pushing the app as a candidate. Derived from the API by default. |
| `vars` | [Vars](#vars) | Environment variables injected into the CF app environment. |
| `deploy_artifact` | string | Path to a file or directory saved by a previous task to deploy to CF. |
| `pre_promote` | Task[] | Tasks to run after the candidate is deployed but before it is promoted to live. TEST_ROUTE is injected. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `pre_start` | string[] | CF CLI commands to run immediately before the candidate app is started. |
| `rolling` | boolean | Use rolling deployment instead of blue-green. |
| `stop_candidate_on_failure` | boolean | Stop the candidate app if deployment fails. |
| `cli_version` | string | CF CLI version to use. Allowed values: cf7, cf8. Defaults to cf7. |
| `docker_tag` | string | Docker image tag to deploy. Required when deploying a Docker image: version or gitref. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |
| `sso_route` | string | Route to configure with SSO. |
| `github_environment` | [GitHubEnvironment](#githubenvironment) | GitHub environment to associate with this deployment. |

### `deploy-katee`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `vars` | [Vars](#vars) | Environment variables available to the vela manifest. |
| `vela_manifest` | string | Path to the vela manifest. Defaults to vela.yaml. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `tag` | string | Deprecated: no longer used - safe to delete. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |
| `environment` | string | **Deprecated.** Deprecated: no longer used - safe to delete. |
| `namespace` | string | Vela namespace to deploy to. Defaults to katee-<team>. |
| `deployment_check_timeout` | integer | **Deprecated.** Deprecated: use max_checks and check_interval instead. |
| `check_interval` | integer | Seconds between each deployment status check. Defaults to 2. |
| `max_checks` | integer | Maximum number of status checks before the deployment is considered failed. Defaults to 60. |
| `github_environment` | [GitHubEnvironment](#githubenvironment) | GitHub environment to associate with this deployment. |

### `deploy-ml-modules`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `ml_modules_version` | string | Version of the ml-modules artifact in Artifactory. |
| `app_name` | string | App name in MarkLogic. Defaults to the pipeline name. |
| `app_version` | string | App version in MarkLogic. Defaults to the git revision. Cannot be set with use_build_version. |
| `targets` | string[] | MarkLogic instances to deploy to. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `use_build_version` | boolean | Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version. |
| `username` | string | Username to connect to MarkLogic. Defaults to the shared vault secret. |
| `password` | string | Password to connect to MarkLogic. Defaults to the shared vault secret. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |

### `deploy-ml-zip`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `deploy_zip` | string | Path to the zip file containing XQuery source files, relative to the manifest. |
| `app_name` | string | App name in MarkLogic. Defaults to the pipeline name. |
| `app_version` | string | App version in MarkLogic. Defaults to the git revision. Cannot be set with use_build_version. |
| `targets` | string[] | MarkLogic instances to deploy to. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `use_build_version` | boolean | Use $BUILD_VERSION instead of $GIT_REVISION for the app version. Cannot be set with app_version. |
| `username` | string | Username to connect to MarkLogic. Defaults to the shared vault secret. |
| `password` | string | Password to connect to MarkLogic. Defaults to the shared vault secret. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |

### `docker-compose`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `command` | string | Command to run against the service. If omitted the default service command is used. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `vars` | [Vars](#vars) | Environment variables available to docker-compose. |
| `service` | string | Name of the docker-compose service to run. Defaults to app. |
| `compose_file` | string | Space-separated list of docker-compose files |
| `save_artifacts` | string[] | Paths to files or directories to save for use in subsequent tasks. |
| `restore_artifacts` | boolean | Restore artifacts saved by previous tasks. |
| `save_artifacts_on_failure` | string[] | Paths to save when the task fails, useful for test reports. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |

### `docker-push`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `username` | string | Username for the target Docker registry. |
| `password` | string | Password for the target Docker registry. |
| `image` | string | Docker image to build and push. Recommended format: eu.gcr.io/halfpipe-io/<team>/<image-name>. |
| `ignore_vulnerabilities` | boolean | Do not fail the build if critical vulnerabilities are found during image scanning. |
| `scan_timeout` | integer | Number of minutes a Trivy vulnerability scan is allowed to run before timing out. |
| `vars` | [Vars](#vars) | Docker build-time variables (ARGs). Do not use for secrets - values are visible in docker history. |
| `secrets` | [Vars](#vars) | Docker build-time secrets, mounted securely during build. |
| `restore_artifacts` | boolean | Restore artifacts saved by previous tasks. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `dockerfile_path` | string | Path to the Dockerfile, relative to the manifest. Defaults to Dockerfile. |
| `build_path` | string | Path to the folder to use as the Docker build context, relative to the manifest. |
| `tag` | string | Deprecated: no longer used - safe to delete. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |
| `platforms` | string[] | Target platforms to build for, e.g. linux/amd64, linux/arm64. Defaults to linux/amd64. |
| `use_cache` | boolean | Enable layer caching to speed up builds by reusing layers from previous builds. |

### `parallel`

| Field | Type | Description |
|-------|------|-------------|
| `tasks` | Task[] | Tasks to run in parallel. All tasks start simultaneously; the group succeeds when all complete. |

### `run`

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional display name. |
| `manual_trigger` | boolean | Task must be manually triggered (Concourse only). |
| `script` | string | Path to the script to execute, relative to the manifest. Prefix with \ to run a system command e.g. \make. |
| `docker` | [Docker](#docker) | Docker configuration for the task to run in. |
| `privileged` | boolean | Run the task as root. Not recommended but sometimes necessary e.g. for docker-in-docker. |
| `vars` | [Vars](#vars) | Environment variables available to the script. |
| `save_artifacts` | string[] | Paths to files or directories to save for use in subsequent tasks. |
| `restore_artifacts` | boolean | Restore artifacts saved by previous tasks. |
| `save_artifacts_on_failure` | string[] | Paths to save when the task fails, useful for test reports. |
| `retries` | integer | Number of times to retry the task if it fails. |
| `notify_on_success` | boolean | **Deprecated.** Deprecated: use notifications instead. |
| `notifications` | [Notifications](#notifications) | Notification channels for this task. |
| `timeout` | string | Timeout duration for the task. If exceeded the task fails. Defaults to 1h. |
| `build_history` | integer | Number of build logs to retain. Defaults to 20 (Concourse only). |

### `sequence`

| Field | Type | Description |
|-------|------|-------------|
| `tasks` | Task[] | Tasks to run in sequence within a parallel group. Can only be used inside a parallel task. |


## Supporting Types

### Notifications

| Field | Type | Description |
|-------|------|-------------|
| `on_success` | string[] | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `on_success_message` | string | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `on_failure` | string[] | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `on_failure_message` | string | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `success` | [NotificationChannels](#notificationchannel) | Notification channels to notify on task success. |
| `failure` | [NotificationChannels](#notificationchannel) | Notification channels to notify on task failure. |

### NotificationChannel

| Field | Type | Description |
|-------|------|-------------|
| `slack` | string | **Deprecated.** Deprecated: Slack notifications are no longer supported. |
| `teams` | string | Microsoft Teams channel webhook URL. |
| `message` | string | Optional message to include in the notification. |

### Vars

Key-value pairs of environment variables. Values are coerced to strings.

```yaml
vars:
  FOO: bar
  PORT: 8080
  DEBUG: true
```

### Docker

Docker image configuration used by the [`run`](#run) task.

| Field | Type | Description |
|-------|------|-------------|
| `image` | string | Path to docker image |
| `username` | string | Username for private Docker registries. |
| `password` | string | Password for private Docker registries. |

### GitHubEnvironment

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Name of the GitHub environment to deploy to. |
| `url` | string | URL associated with the GitHub environment. |

### Feature Toggles

Available values for the `feature_toggles` array:

- `update-pipeline`
- `update-pipeline-and-tag`
- `github-statuses`
- `ghas`

