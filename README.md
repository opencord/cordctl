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
cordctl -h

# Show help for model-related commands
cordctl model -h

# Show help for the model list command
cordctl model list -h
```

## Shell Completion
`cordctl` supports shell completion for the `bash` shell. To enable
shell Completion you can use the following command on *most* \*nix based system.
```bash
source <(cordctl completion bash)
```

If this does not work on your system, as is the case with the standard
bash shell on MacOS, then you can try the following command:
```bash
source /dev/stdin <<<"$(cordctl completion bash)"
```

If you which to make `bash` shell completion automatic when you login to
your account you can append the output of `cordctl completion bash` to
your `$HOME/.bashrc`:
```bash
cordctl completion bash >> $HOME/.bashrc
```

## Interacting with models

`cordctl` has several commands for interacting with models in XOS:

* `cordctl modeltype list` ... list the types of models that XOS supports.
* `cordctl model list <modelName>` ... list instances of the given model, with optional filtering.
* `cordctl model update <modelName> <id> --set-json <json>` ... update models with new fields

### Listing model types

XOS supports a dynamic set of models that may be extended by services that are loaded into XOS. As such the set of models that XOS supports is not necessarily fixed at initial deployment time, but may evolve over the life cycle of an XOS deployment as services are added, removed, or upgraded. The `modeltype list` command allows you to query the set of model types that XOS supports. For example,

```bash
# Query available model types
cordctl modeltype list
```

### Listing models

The basic syntax for listing models (`cordctl model list <modelName>`) will list all objects of a particular model. Filtering options can be added by using the `--filter` argument and providing a comma-separated list of filters. If the filters contain characters that the shell would interpret, such as spaces or `>`, `<` or `!` then you'll need to escape your filter specifier. For example,

```bash
# List slices that have id > 10 and controller_kind = Deployment
cordctl model list Slice --filter "id>10, controller_kind=Deployment"
```

Supported operators in the filters include `=`, `!=`, `>`, `<`, `>=`, `<=`.

### Updating models

The `model update` command is a flexible way to update one or more models. The most basic syntax uses one or more model IDs. For example,

```bash
# Update Site 1 and set its site_url to http://www.opencord.org/
cordctl model update Site 1 --set-field site_url=http://www.opencord.org/
```

Alternatively you may specify a JSON-formatted dictionary. Make sure to properly quote your JSON dictionary when using it as a command line argument. For example,

```bash
# Update Site 1 and set its site_url to http://www.opencord.org/
cordctl model update Site 1 --set-json '{"site_url": "http://www.opencord.org/"}'
```

If you don't know the ID of the object you wish to operate, or if you want to update several objects at the same time that have something in common, then you can use a `--filter` argument instead of an ID. For example,

```bash
# Update all sites named "mysite"  and set its site_url to http://www.opencord.org/
cordctl model update Site --filter name=mysite --set-field site_url=http://www.opencord.org/
```

## Development Environment

To run unit tests, `go-junit-report` and `gocover-obertura` tools must be installed. One way to do this is to install them with `go get`, and then ensure your `GOPATH` is part of your `PATH` (editing your `~/.profile` as necessary). 

```bash
go get -u github.com/jstemmer/go-junit-report
go get -u github.com/t-yuki/gocover-cobertura
```

