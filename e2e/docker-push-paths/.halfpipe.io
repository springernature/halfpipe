team: test
pipeline: test

triggers:
- type: git
  watched_paths:
  - e2e/docker-push-paths

tasks:
- type: docker-push
  name: push to docker registry
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe-fly
  dockerfile_path: dockerfiles/Dockerfile
  build_path: some/build/dir

- type: docker-push
  name: push to docker registry again
  username: rob
  password: verysecret
  image: springerplatformengineering/halfpipe
  dockerfile_path: dockerfiles/Dockerfile
