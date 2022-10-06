#!/bin/bash

docker build . -t signing-agent-testrunner
docker run -ti -v $PWD/../..:/src/gitlab.qredo.com/custody-engine/signing-agent signing-agent-testrunner sh
