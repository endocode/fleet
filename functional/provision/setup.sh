#!/bin/bash -e

USER_ID=${SUDO_UID:-$(id -u)}
HOME=$(getent passwd "${USER_ID}" | cut -d: -f6)

export GOROOT=${HOME}/go
export PATH=${HOME}/go/bin:${PATH}

if [ ! -f ${HOME}/go1.5.3.linux-amd64.tar.gz ]; then
  # Remove unfinished archive when you press Ctrl+C
  trap "rm -f ${HOME}/go1.5.3.linux-amd64.tar.gz" INT TERM
  wget --no-verbose https://storage.googleapis.com/golang/go1.5.3.linux-amd64.tar.gz -P ${HOME}
fi
tar -xf ${HOME}/go1.5.3.linux-amd64.tar.gz -C ${HOME}
