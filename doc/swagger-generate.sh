#!/bin/bash
SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -o doc/swagger/swagger.yaml \
      -c rest                                                                     \
      -c gitlab.qredo.com/custody-engine/signing-agent/api
#      -c doc