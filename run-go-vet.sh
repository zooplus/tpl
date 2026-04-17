#!/usr/bin/env bash
set -e
pkg=github.com/zooplus/tpl
for dir in ; do
  go vet /
done
