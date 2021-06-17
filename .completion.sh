#!/usr/bin/env bash

for s in bash zsh fish powershell; do
  for cmd in mcp mcs mcx mec mgo mping mtun; do
    ${cmd} --completion ${s} > docs/completion/${cmd}.${s}
  done
done