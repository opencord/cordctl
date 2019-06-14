# cordctl

---

`cordctl` is a command-line tool for interacting with XOS. XOS is part of the SEBA NEM and part of CORD, so by extension this tool is useful for interacting with SEBA and CORD deployments. `cordctl` makes use of gRPC to connect to XOS and may by used for administration of a remote pod, assuming the appropriate firewall rules are configured. Typically XOS exposes its gRPC API on port `30011`.

## Obtaining cordctl

Binaries for `cordctl` are published at https://github.com/opencord/cordctl/releases and may be directly downloaded and used on Linux, MAC, or Windows platforms.

Additionally, the source for `cordctl` is available at https://github.com/opencord/cordctl and may be downloaded and built. `cordctl` is written in golang, and go version 1.12 or above must be installed in order to build from source.

If you would like to contribute to `cordctl`, the preferred process is to submit patches for code review through gerrit at https://gerrit.opencord.org.

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

To generate a config file on stdout from the currently configured settings, the command `cordctl config` may be used.

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
* `cordctl model update <modelName> <id> --set-json <json>` ... update models with new fields.
* `cordctl model delete <modelName> <id>` ... delete models.
* `cordctl model create <modelName> --set-json <json>` ... create a new model.
* `cordctl model sync <modelName> <id>` ... wait for a model to be synchronized.
* `cordctl model setdirty <modelName> <id>` ... set a model dirty so it will be synchronized.

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

The core also permits models to be filtered based on state, and the `--state` argument can be used to filter based on a state. States include `all`, `dirty`, `deleted`, `dirtypol`, and `deletedpol`. `default` is a synonym for `all`. For example,

```bash
# List deleted ONOSApps
cordctl model list ONOSApp --state deleted
```

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

### Deleting Models

The syntax for deleting models is similar to that for updating models. You may delete by specifying one of more IDs, or you may delete by using a filter. For example,

```bash
# Delete Slice 1
cordctl model delete Slice 1

# Delete the Slice named myslice
cordctl model delete Slice --filter name=mylice
```

### Creating Models

The `model create` command allows you to create new instances of a model in XOS. To do this, specify the type of the model that you want to create and the set of fields that populate it. The set of fields can be set using a name=value syntax or by passing a JSON object. The following two examples are equivalent,

```base
# Create a Site by using set-field
cordctl model create Site --set-field name=somesite,abbreviated_name=somesite,login_base=somesite

# Create a Site by passing a json object
cordctl model create Site --set-json '{"name": "somesite", "abbreviated_name": "somesite", "login_base": "somesite"}'
```

### Syncing Models

All XOS operations are by nature asynchronous. When a model instance is created or updated, a synchronizer will typically run at a later time, enacting that model by applying its changes to some external component. After this is complete, the synchronizer updates the timestamps and other metadata to convey that the synchronization is complete.

Asynchronous operations are often inconvenient for test infrastructure, automation scripts, and even human operators. `cordctl` offers some features for synchronous behavior.

The first is the `model sync` command that can sync models based on ID or based on a filter. For example,

```bash
# Sync based on ID
cordctl model sync ONOSApp 17

# Sync based on a field filter
cordctl model sync ONOSApp --filter name=olt
```

The second way to sync an object is to use the `--sync` command when doing a `model create` or a `model update`. For example,

```bash
cordctl model create ONOSApp --sync --set-field name=myapp,app_id=org.opencord.myapp
```

### Dirtying Models

XOS determines when a model is dirty by comparing timestamps. Each model has an `updated` timestamp that indicates the last time the model was
updated. If this timestamp is newer than the `enacted` or `policed` timestamps then sync steps or policies will be run respectively. There is
no direct way to modify the `updated` timestamp via the API since timestamps are managed by the `XOS` core. `cordctl` provides the `setdirty` command to cause models to be dirtied without altering the other fields of the model. When `setdirty` is used, the `updated` timestamp will be set to the current time. The `setdirty` command may be used with either an ID or a filter or the `--all` flag.

```bash
# Set model dirty based on ID
cordctl model setdirty ONOSApp 17

# Set model dirty based on a field filter
cordctl model setdirty ONOSApp --filter name=olt

# Set all ONOSApp models dirty
cordctl model setdirty ONOSApp --all
```

If you wish to query which models are dirty (for example, to verify that a previous `setdirty` worked as expected) then the `--state dirty` argument may be applied to the `model list` command. For example,

```bash
# Get a list of dirty ONOS Apps
cordctl model list ONOSApp --state dirty
```

> Note: Not all models have syncsteps or policies implemented for them. Some models may implicitly cause related models to become
> dirty. For example, dirtying the head of a service instance chain may cause the whole chain to be dirtied. This behavior is dependent on
> the model. For models that do not implement syncsteps or policies and do not have the side-effect of dirtying related models, the `setdirty`
> command has no practical value, and the models may remain in perpetual dirty state.

## Development Environment

To run unit tests, `go-junit-report` and `gocover-obertura` tools must be installed. One way to do this is to install them with `go get`, and then ensure your `GOPATH` is part of your `PATH` (editing your `~/.profile` as necessary).

```bash
go get -u github.com/jstemmer/go-junit-report
go get -u github.com/t-yuki/gocover-cobertura
```

