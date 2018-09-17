# 2. Task Cache Volumes

Date:

## Status

Proposed

## Context

Concourse has [task caches](https://concourse-ci.org/tasks.html#task-caches) to save state between runs of the same task on the same worker.


## Decision

Change halfpipe to provide one directory `/halfpipe-cache` instead of a list of directories specific to common build tools.

This will allow users to configure any build tool to use the cache instead of the onus being on halfpipe to add support.

Also there is a small overhead to mounting each cache volume, so one volume is better than n.


## Consequences

Users will have to point their build tool  at `/halfpipe-cache` instead of `/root/.sbt` for example.



