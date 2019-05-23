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
for i in $(find /xos/v1 -name "*.proto"); do
    sed -i -e 's/\[(child_node) = {}\]//g' $i
done
