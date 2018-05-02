# Hello!

This is the friendly Halfpipe CLI. Try it out :)

# Friendly Halfpipe CLI huh?

Yeah, it takes a small YAML schema and renders a complete Concourse pipeline for you

# How does it work? This README is kinda slim...

[All documentation and further information can be found here](https://docs.halfpipe.io)

# Ah! Thats cool, can I use it in my company?

In theory yes, but there is some Springer Nature specific stuff in here. With that said nothing is stopping us from extracting those bits, submit a issue! :)

# How to I test and build?

```
go get github.com/springernature/halfpipe
go get golang.org/x/tools/cmd/goimports
go get -u github.com/alecthomas/gometalinter && gometalinter --install
brew install dep

cd $GOPATH/src/github.com/springernature/halfpipe
./build.sh
```
