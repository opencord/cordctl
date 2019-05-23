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
# See the License for the specific language governing permissions and
# limitations under the License.

# docker build -t xosproject/mock-xos:candidate .

FROM xosproject/grpc-mock:1.0.0 as build
MAINTAINER Open Networking Foundation

RUN apk add --update git python2 py-pip make

WORKDIR /xos
RUN mkdir /xos/v1
RUN git clone https://github.com/opencord/xos /opt/tmp_xos && mv /opt/tmp_xos/xos /opt/xos && mv /opt/tmp_xos/lib /opt/xos/lib && cp /opt/tmp_xos/VERSION /opt/xos
RUN cd /opt/xos/lib/xos-genx && python setup.py install
RUN cd /opt/xos/coreapi/protos && make rebuild-protos

RUN cp -r /opt/xos/coreapi/protos/*.proto /xos/v1

COPY clean.sh /xos/clean.sh
RUN /xos/clean.sh

FROM xosproject/grpc-mock:1.0.0
ENV GRPC_MOCK_COMPARE=sparse

WORKDIR /xos
COPY --from=build /xos /xos
COPY mock.sh clean.sh mock-v1.js data.json /xos/
RUN chmod 755 /xos/mock.sh /xos/clean.sh

ENTRYPOINT ["/xos/mock.sh"]
