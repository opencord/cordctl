#!/usr/bin/env sh
# Portions copyright 2019-present Open Networking Foundation
# Original copyright 2019-present Ciena Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

MOCK_SCRIPT=/xos/mock-v1.js

while getopts ":12h" opt; do
  case ${opt} in
    1)
      MOCK_SCRIPT=/xos/mock-v1.js
      ;;
    h)
      echo "usage: $PROG [-1|-2] [-h]"
      echo "  -1    Mock version 1 XOS server"
      echo "  -h    Display this message"
      exit 0
      ;;
    *)
      echo "usage: $PROG [-1|-2] [-h]"
      echo "  -1    Mock version 1 XOS server"
      echo "  -h    Display this message"
      exit 1
      ;;
  esac
done

node $MOCK_SCRIPT
