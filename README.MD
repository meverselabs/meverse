This readme file is deprecated. It will be updated soon.

## FLETA

This is repository of FLETA project.<br/>
This repository contains an explanation and contents about the FLETA Daemon, Observer Daemon, Block Explorer, and DApp examples.<br/>
This v2 of FLETA blockchain provides more easy and convenient way to develop DApp.<br/>

## Related repositories
* You can find the Explorer at https://github.com/fletaio/explorer.

## Building the source

### Prerequisites
Golang 1.18.2 or later<br/>
C compiler(gcc 4.9.0 or later, mingw etc.)<br/>

### Step
Open the Command line and enter the command below.

Build Node(default, join to the testnet)
```
$ git clone https://github.com/meverselabs/meverse
$ cd ./meverse/cmd/node
$ go build
$ ./node
```

Build Generator(required to create generator account for joining to the testnet as a observer)
```
$ git clone https://github.com/meverselabs/meverse
$ cd ./meverse/cmd/generator
$ go build
$ ./formulator
```

Build Observer(buildable, but cannot join to the testnet as a observer)
```
$ git clone https://github.com/meverselabs/meverse
$ cd ./meverse/cmd/observer
$ go build
$ ./observer
```

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

All the requirements mentioned in this document are written based on MEVerse’s full node. All requirements can be changed at the release stage.
Currently, we can achieve 7,000 TPS with 36 6-core servers in global regions, and now we are developing our system to achieve 20,000 TPS, which is the objective of the project.
##
### Configuration

In current stage, only the minimal configuration for the execution is being provided.<br/>
As the development is progressed, there can be additions/deletions/changes at any time.<br/>
Please see the details in related repositories.

## Running on FLETA Beta Test network

Just running node will join to the test network.<br/>
If you want to join as a formulator, you should make formulator account.<br/>
It will be provided soon.<br/>

## Running on your own network

It will be provided soon.<br/>

## License

All codes under this repository are licensed under the [GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also included in our repository in the `LICENSE` file.
