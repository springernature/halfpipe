team: halfpipe-team
pipeline: halfpipe-e2e-copy-container-image

tasks:
  - type: copy-container-image
    name: cp
    source: eu.gcr.io/halfpipe-io/team/image:tag
    target: 1234567890.dkr.ecr.cn-northwest-1.amazonaws.com.cn
