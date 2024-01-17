#!/bin/bash

set -x

cmd="make test"
if [ $# -gt 0 ]; then
    cmd="$@"
fi

fswatch -l 0.2 -e '\.#' -e '#' -e '~' -e '/bin/' -e worksets -e log . | xargs -n1 sh -c \
    "echo  && echo =======\> EXEC; $cmd && echo && date && echo =======\> RESULT \$?"
