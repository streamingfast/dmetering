#!/bin/bash -u

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )"

function main() {
  checks

  set -e

  pushd "$ROOT" >/dev/null
    buf generate proto
  popd >/dev/null

  echo "generate.sh - `date` - `whoami`" > $ROOT/pb/last_generate.txt
  echo "Done"
}

function checks() {
  result=`buf --version 2>&1 | grep -Eo 1.[2-9][0-9]+`
  if [[ "$result" == "" ]]; then
    echo "Your version of 'buf' (at `which buf` with version `buf --version`) is not recent enough."
    echo ""
    echo "To fix your problem, install latest 'buf' CLI:"
    echo ""
    echo "  https://buf.build/docs/installation"
    echo ""
    echo "If everything is working as expetcted, the command:"
    echo ""
    echo "  buf --version"
    echo ""
    echo "Should print '1.30.0' (output differs)"
    exit 1
  fi
}

main "$@"
