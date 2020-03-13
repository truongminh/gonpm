#!/bin/sh
# clean
rm -rf false
rm -rf node_modules

PROXY=http://localhost:8999
REGISTRY=https://registry.npmjs.org
OPTIONS="--proxy $PROXY --https-proxy $PROXY --registry $REGISTRY --no-audit --no-cache --strict-ssl false --no-package-lock"
npm $OPTIONS install is-positive array-first

node index.js
