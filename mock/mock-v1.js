// Copyright 2019 Ciena Corporation
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
const mockServer = createMockServer({
  protoPath: "utility.proto",
  packageName: "xos",
  serviceName: "utility",
  options: {
          keepCase: true,
          enum: "String",
          defaults: true,
          oneofs: true,
          includeDirs: ["/xos/v1"],
          address: "0.0.0.0:50051"
  },
  rules: require('/xos/data.json')
});

/*
mockServer.use({  
  protoPath: "xos.proto",
  packageName: "xos",
  serviceName: "xos",
  options: {
        keepCase: true,
        enum: "String",
        defaults: true,
        oneofs: true,
        includeDirs: ["/xos/v1"],
        address: "0.0.0.0:50051"
  }},
  mockServer.routes)
  */

process.on('SIGINT', function() {
    console.log("Caught interrupt signal");
    process.exit();
});
mockServer.listen("0.0.0.0:50051");
