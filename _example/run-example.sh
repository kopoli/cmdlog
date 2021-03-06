#!/bin/sh

usage() {
    cat <<EOF
Usage: $0

Run an example zsh session with cmdlog.

EOF
}

if ! type zsh >/dev/null 2>&1; then
    echo "This example requires zsh to be installed and in PATH."
    exit 1
fi

env --ignore-environment ZDOTDIR="$PWD" \
    PATH="$PATH" \
    CMDLOG_FILE="$PWD/cmdlog.example.log" \
    CMDLOG_FILTERS="$PWD/cmdlog-filters.example.txt" \
    PS1="cmdlog-test> " \
    TERM="$TERM" \
    zsh --no-globalrcs
