#!/usr/bin/env bash

# put something like this in your ~/.bash_profile
# [ -f ~/go/src/github.com/springernature/halfpipe/extras/halfpipe-helpers.sh ] && source ~/go/src/github.com/springernature/halfpipe/extras/halfpipe-helpers.sh



fly-login() {
    team="${1:-"engineering-enablement"}"
    target="${team}"
    if [[ "${team}" == "engineering-enablement" ]]; then
      target=ee
    fi
    fly -t ${team} login \
      -c https://concourse.halfpipe.io \
      -n ${team} \
      -u "$(vault read -field=username springernature/${team}/concourse)" \
      -p "$(vault read -field=password springernature/${team}/concourse)"
}
