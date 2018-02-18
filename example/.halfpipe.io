team: engineering-enablement

repo:
  uri: https://github.com/springernature/foo.git
  private_key: rrrr

tasks:
- name: run
  script: ./test.sh
  image: busybox

- name: run
  script: ./build.sh
  image: busybox

- name: deploy-cf
  api: dev
  space: dev

- name: deploy-cf
  api: live
  space: staging

- name: deploy-cf
  api: live
  space: live

- name: deploy-cf
  api: https://some.custom.cf
  space: live

- name: run
  script: ./notify.sh
  image: busybox

- name: docker-push
  username: user1
  password: pass1
  repo: foo/bar
