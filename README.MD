## FLETA

This is repository of FLETA project.<br/>
This repository contains an explanation and contents about the FLETA Daemon, Observer Daemon, Block Explorer, and DApp examples.<br/>
This v2 of FLETA blockchain provides more easy and convenient way to develop DApp.<br/>
#### !!!Currently, we are developing FLETA blockchain v2 that reworks whole structure!!!
#### !!!Currently, consensus by nodes on the network is remained to reworks(already implemented at v1, but conversion required)!!!
#### !!!Currently, example and explorer is running on single node by local consensus. It will be removed when we finish to convert our consensus by node on the network!!!

## Related repositories
* You can find the Explorer at https://github.com/fletaio/explorer.

## Building the source

### Prerequisites
Golang 1.11 or later<br/>
C compiler(gcc, mingw etc.)<br/>

### Step
Open the Command line and enter the command below.

```
$ git clone https://github.com/fletaio/fleta
$ cd ./fleta
$ go build
$ ./node.exe
```

You can find FLETA binary in GOPATH/bin directory

## Environment

### Command line and flags
Configuration and specification are continuously changing, and currently, all configurations are provided via configuration file.<br/>
When the changing possibility of configurations and command sets are reduced, command line and flags will be provided.<br/>

### System requirements

| Resource | Recommended | Minimum |
|:---------:|---------------|---------|
|CPU|6 core for 10,000 TPS|Minimum 2 core for 3,000 TPS|
|RAM|4 GB|Minimum 1 GB|
|Disk|16 GB|Minimum 4 GB|

All the requirements mentioned in this document are written based on FLETA’s full node. All requirements can be changed at the release stage.
Currently, we can achieve 10,000 TPS with 36 6-core servers in global regions, and now we are developing our system to achieve 20,000 TPS, which is the objective of the project.
##
### Configuration

In current stage, only the minimal configuration for the execution is being provided.<br/>
As the development is progressed, there can be additions/deletions/changes at any time.<br/>
Please see the details in related repositories.

## Running on FLETA Beta Test network

It will be provided when we finish to convert our consensus on nodes on the network.<br/>

## Running on your own network

It will be provided when we finish to convert our consensus on nodes on the network.<br/>

## License

All codes under this repository are licensed under the [GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also included in our repository in the `LICENSE` file.
