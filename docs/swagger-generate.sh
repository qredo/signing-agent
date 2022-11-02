#!/bin/bash
SWAGGER_GENERATE_EXTENSION=false swagger generate spec -m -o docs/swagger/swagger.yaml \
      -c rest                                                                     \
      -c gitlab.qredo.com/computational-custodian/signing-agent/api
#      -c doc