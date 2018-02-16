team: asd

repo:
  uri: https://github.com/springernature/foo.git


tasks:
- name: run
  script: ./test.sh
  image: busybox

- name: run
  script: ./build.sh
  image: busybox

- name: deploy-cf
  api: sdpokasd
  username: uname
  password: pwd
  org: orgname
  space: spacename

- name: deploy-cf
  api: sdpokasd2
  username: uname2
  password: pwd2
  org: orgname2
  space: spacename2

- name: run
  script: ./notify.sh
  image: busybox

- name: docker-push
  username: user1
  password: pass1
  repo: foo/bar
