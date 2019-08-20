team: test
pipeline: test

triggers:
- type: git
  watched_paths:
  - e2e/consumer-integration-test

tasks:
- type: consumer-integration-test
  name: another-c-name
  consumer: c-consumer
  consumer_host: c-host
  provider_host: p-host
  script: c-script
  docker_compose_service: potato
  git_clone_options: "--depth 100"
  vars:
    K: value
    K1: value1
