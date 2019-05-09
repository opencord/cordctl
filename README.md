# cordctl
---

`cordctl` is a command-line tool for interacting with XOS. XOS is part of the SEBA NEM and part of CORD, so by extension this tool is useful for interacting with SEBA and CORD deployments. `cordctl` makes use of gRPC to connect to XOS and may by used for administration of a remote pod, assuming the appropriate firewall rules are configured. Typically XOS exposes its gRPC API on port `30011`.

## Configuration

Typically a configuration file should be placed at `~/.cord/config` as `cordctl` will automatically look in that location. Alternatively, the `-c` command-line option may be used to specify a different config file location. Below is a sample config file:

```yaml
server: 10.201.101.33:30011
username: admin@opencord.org
password: letmein
grpc:
  timeout: 10s
```

The `server`, `username`, and `password` parameters are essential to configure access to the XOS container running on your pod. 

## Getting Help

The `-h` option can be used at multiple levels to get help, for example:

```bash
# Show help for global options
./cordctl -h

# Show help for model-related commands
./cordctl model -h

# Show help for the model list command
./cordctl model list -h
```

## Development Environment

To run unit tests, `go-junit-report` and `gocover-obertura` tools must be installed. One way to do this is to install them with `go get`, and then ensure your `GOPATH` is part of your `PATH` (editing your `~/.profile` as necessary). 

```bash
go get -u github.com/jstemmer/go-junit-report
go get -u github.com/t-yuki/gocover-cobertura
```

