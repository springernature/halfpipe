teamx: engineering-enablement
repo:
  uri: https://zzz
  private_key: asdf
tasks:
- name: run
  script: ./test.sh
  image: openjdk:8-slim
- name: docker-push
  username: ((docker.username))
  password: ((docker.password))
  repository: simonjohansson/half-pipe-linter
- name: deploy-cf
  space: test
  api: https://api.europe-west1.cf.gcp.springernature.io
- name: run
  script: ./asdf.sh
  image: openjdk:8-slim
  vars:
    A: asdf
    B: 1234
- name: deploy-cf
  space: test
  api: https://api.europe-west1.cf.gcp.springernature.io
  vars:
    VAR1: asdf1234
    VAR2: ((a.secret.in.vault))
