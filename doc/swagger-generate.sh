#!/bin/bash
SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -o doc/swagger/swagger.yaml \
      -c rest                                                                     \
      -c gitlab.qredo.com/qredo-server/core-client/api
#      -c doc