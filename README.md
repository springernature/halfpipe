# Hello!

<a href="https://concourse.halfpipe.io/teams/engineering-enablement/pipelines/halfpipe-cli"><img src="http://badger.halfpipe.io/engineering-enablement/halfpipe-cli" title="badge"></a>

This is the friendly Halfpipe CLI. Try it out :)

## Friendly Halfpipe CLI huh?

Yeah, it takes a small YAML schema and renders a complete Concourse pipeline for you

## How does it work? This README is kinda slim...

[All documentation and further information can be found here](https://docs.halfpipe.io)

## Ah! That's cool, can I use it in my company?

In theory yes, but there is some Springer Nature specific stuff in here. With that said nothing is stopping us from extracting those bits, submit a issue! :)

## How do I test and build?

Halfpipe is built with Go

```
./build.sh

# or if you don't have Go installed you can build in docker
./docker-build.sh
```

# CI

The main pipeline is in [Concourse](https://concourse.halfpipe.io/teams/engineering-enablement/pipelines/halfpipe-cli)

It runs the build script on every commit to `main`.

We also use [GitHub Actions](https://github.com/springernature/halfpipe/actions) for dependabot and CodeQL scanning


# Updating Dependencies


### go

dependabot will raise PRs. Alternatively, to manually update all deps:

```bash
go get -t -u ./...
go mod tidy
```

### GitHub actions 

For third party actions we use in halfpipe rendered workflows - 
dependabot will raise PRs but these are just informational, we have to manually update the halfpipe actions renderer. 


# Releasing

Releasing is triggered by manually bumping the version (major, minor or patch) in Concourse. Binaries are built for different platforms and published to artifactory. The halfpipe cli checks artifactory for a newer release and updates itself. A [GitHub release](https://github.com/springernature/halfpipe/releases) is also created.
