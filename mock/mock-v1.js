// Portions copyright 2019-present Open Networking Foundation
// Original copyright 2019-present Ciena Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
const {createMockServer} = require("grpc-mock");
const options = {
  keepCase: true,
  enum: "String",
  defaults: true,
  oneofs: true,
  includeDirs: ["/xos/v1"],
  address: "0.0.0.0:50051"}
const mockServer = createMockServer({
  protoPath: "xos.proto",
  packageName: "xos",
  serviceName: "xos",
  options: options,
  rules: require('/xos/data.json')
});

// add xos.utility protos to server
mockServer.addProtos({
  protoPath: "utility.proto",
  packageName: "xos",
  serviceName: "utility",
  options: options});

// add xos.dynamicload protos to server
mockServer.addProtos({
  protoPath: "dynamicload.proto",
  packageName: "xos",
  serviceName: "dynamicload",
  options: options});

// add xos.filetransfer protos to server
mockServer.addProtos({
  protoPath: "filetransfer.proto",
  packageName: "xos",
  serviceName: "filetransfer",
  options: options});

process.on('SIGINT', function() {
    console.log("Caught interrupt signal");
    process.exit();
});
mockServer.listen("0.0.0.0:50051");
console.log("Listening for requests\n")
