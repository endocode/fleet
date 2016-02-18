#!/bin/sh

USER_ID=${SUDO_UID:-$(id -u)}
USER=$(getent passwd "${USER_ID}" | cut -d: -f1)
HOME=$(getent passwd "${USER_ID}" | cut -d: -f6)

export GOROOT=${HOME}/go
export PATH=${PATH}:${HOME}/go/bin

if [[ ! $(go version 2>/dev/null) ]]; then
  if [ ! -f ${HOME}/go1.5.3.linux-amd64.tar.gz ]; then
    # Remove unfinished archive when you press Ctrl+C
    trap "rm -f ${HOME}/go1.5.3.linux-amd64.tar.gz" EXIT
    wget --progress=dot:mega https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz -P ${HOME}
  fi
  tar -xf ${HOME}/go1.5.3.linux-amd64.tar.gz -C ${HOME}
fi

