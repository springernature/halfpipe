# Example project

An example using all halfpipe features.

Run halfpipe in this directory and push to concourse:

```
halfpipe > pipeline.yml
fly -t ci set-pipeline -p testdeploy -c pipeline.yml
`
