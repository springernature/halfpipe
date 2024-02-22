team: halfpipe-team
pipeline: pipeline-name
platform: actions

triggers:
- type: git
  watched_paths:
  - .

tasks:
- type: consumer-integration-test
  name: c-name
  consumer: c-consumer/sub/dir
  consumer_host: c-host
  provider_host: p-host
  provider_name: p-name
  script: c-script
  docker_compose_file: custom-docker-compose.yml
  docker_compose_service: potato
  git_clone_options: "--depth 100"
  vars:
    K: value
    K1: value1
    S1: ((very.secret))

- type: consumer-integration-test
  name: c-name-covenant
  use_covenant: false
  consumer: c-consumer
  consumer_host: c-host
  provider_host: p-host
  provider_name: p-name
  script: c-script
  docker_compose_service: potato
  git_clone_options: "--depth 100"
  vars:
    K: value
    K1: value1
    S1: ((very.secret))
