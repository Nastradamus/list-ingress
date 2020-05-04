#!/usr/bin/env bash
set -euo pipefail

printf "\n*** RUN go fmt ***\n"
unformatted=$(gofmt -l -w -s .)
if [ "$unformatted" ]; then
  printf "unformatted files:\n%s\n" "${unformatted}"
  exit 1
fi

printf "\n*** RUN golangci-lint ***\n"
golangci-lint run

printf "\n*** CODE STYLE IS OK ***\n"
