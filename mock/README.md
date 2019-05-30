This directory contains a mock server that is used when unit-testing the cordctl client. This mock server is implemented in javascript and runs using nodejs inside of a docker container.

Because the nodejs gRPC server implementation does not yet support the gRPC reflection API, a set of statically built protobufs, `xos-core.protoset`, is made available and can be used with the client instead of the reflection API. This protoset is not intended to be an operational substitute for the API retrieved via reflection, but is merely a convenience for unit testing.

To regenerate the protoset from inside a running xos-core container, do this:

```bash
python -m grpc.tools.protoc --proto_path=. --descriptor_set_out=xos-core.protoset --include_imports xos.proto utility.proto filetransfer.proto dynamicload.proto modeldefs.proto
```
