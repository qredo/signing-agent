#!/bin/bash
SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -o docs/swagger/swagger.yaml \
      -c rest                                                                     \
      -c signing-agent/api
#      -c doc