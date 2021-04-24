#!/bin/sh

test -z "$GOPATH" && GOPATH="$HOME/go"

if ! test -x $GOPATH/bin/implant; then
  IMPLANT_GIT="https://github.com/skx/implant"
  (mkdir -p $HOME/git; cd $HOME/git; git clone $IMPLANT_GIT; cd implant; go install)
fi

$GOPATH/bin/implant -verbose -package staticfiles -input files


