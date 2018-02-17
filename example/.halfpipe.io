team: asd

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
  space: spacename1
  username: uname1
  password: pwd1

- name: deploy-cf
  api: https://some.custom.cf
  username: uname2
  password: pwd2
  org: orgname2
  space: spacename2

- name: deploy-cf
  api: live
  username: uname3
  password: pwd3
  space: spacename3

- name: run
  script: ./notify.sh
  image: busybox

- name: docker-push
  username: user1
  password: pass1
  repo: foo/bar
