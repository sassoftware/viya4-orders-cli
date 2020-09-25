# SAS Viya Orders CLI
SAS Viya Orders is a command-line interface (CLI) that calls the appropriate 
[SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html) 
endpoint to obtain the requested deployment assets for a specified SAS Viya 
software order.

## Overview
You can use this CLI both as a tool and as an example of how to use Golang to call the 
[SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html). You can 
also import the assetreqs and authn packages and use them in your own Golang project.
```
Usage:
  viya4-orders-cli [command]

Available Commands:
  certificates     Download certificates for the given order number
  deploymentAssets Download deployment assets for the given order number at the given cadence name and version - if version not specified, get the latest version of the given cadence name
  help             Help about any command
  license          Download a license for the given order number at the given cadence name and version

Flags:
  -c, --config string      config file (default is $HOME/.viya4-orders-cli)
  -n, --file-name string   name of the file where you want the downloaded order asset stored
                           (defaults:
                                certs - SASiyaV4_<order number>_certs.zip
                                license and depassets - SASiyaV4_<order number>_<renewal sequence>_<cadence information>_<asset name>_<date time stamp>.<asset extension>
                           )
  -p, --file-path string   path to where you want the downloaded order asset stored (default is path to your current working directory)
  -h, --help               help for viya4-orders-cli
  -o, --output string      output format - valid values:
                                j, json
                                t, text
                            (default "text")
  -v, --version            version for viya4-orders-cli

Use "viya4-orders-cli [command] --help" for more information about a command.
```
### Prerequisites

* [Go](https://golang.org/) 1.13 or [Docker](https://www.docker.com/) is required.
* API credentials for the [SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html)
are required. You can obtain them from the [SAS API Portal](https://apiportal.sas.com/get-started).

### Installation

**Note:** You can build a Docker image directly from the GitHub repository. Otherwise, 
you must first clone the project before you can use SAS Viya Orders CLI.

To build the project, you have several options:

* Using [Make](https://www.gnu.org/software/make/): <br>
  ```
  make build
  ```

* Using [Docker](https://www.docker.com/): <br>
  * If you have already cloned the project, execute the following from the project 
  root directory:
     ```
     docker build . -t viya4-orders-cli
     ```
  * Or you can build the project without cloning first:
     ```
     docker build github.com/sassoftware/viya4-orders-cli -t viya4-orders-cli
     ```

* Using `go build`: <br>
   ```
   go build -o viya4-orders-cli main.go
   ```

## Getting Started
If you do not yet have credentials for SAS Viya Orders API, obtain them from
  the [SAS API Portal](https://apiportal.sas.com/get-started).

If you intend to use Make or Docker, edit your makefile or dockerfile to add
  the command and order number, plus other arguments that are described in the
  [Overview](#Overview) section.

Select CLI options. You can then specify them on the command line, pass them in
  as environment variables, or include them in a config file.

  SAS Viya Orders CLI options are applied in order of precedence, as follows:
  1. command-line specification
  1. environment variable specification
  1. config file specification
  1. default
  
Base64 encode the OAuth client ID and client secret that serve as your API credentials and define 
them as environment variables. You can also add them 
to your config file as values for `clientCredentialsId` and `clientCredentialsSecret`,
respectively.

If you want to use a config file, create it. The default file is `$HOME/.viya4-orders-cli`.
The config file must be in [YAML](https://yaml.org/) format, or else its file name must 
include a file extension that denotes another format. Supported formats are [JSON](https://www.json.org/), 
[TOML](https://github.com/toml-lang/toml), [YAML](https://yaml.org/), [HCL](https://github.com/hashicorp/hcl), 
[INI](https://docs.microsoft.com/en-us/previous-versions/windows/desktop/ms717987(v=vs.85)), 
[envfile](https://www.npmjs.com/package/envfile) or 
[Java properties](https://docs.oracle.com/javase/tutorial/essential/environment/properties.html) formats.

### Running

You have the following options for launching SAS Viya Orders CLI:

* Using [Make](https://www.gnu.org/software/make/): <br>
   ```unix
   make run
   ```

* Using [Docker](https://www.docker.com/) (assumes you have already executed the Docker build step described
 in the [Installation section](#installation)): <br>
   ```docker
   docker run viya4-orders-cli
   ```

* Using `go run`: <br>
   ```
   go run main.go [command] [args] [flags]
   ```

### Examples

The examples in this section correspond to typical tasks that you might 
perform using SAS Viya Orders API:

* Get deployment assets for the initial deployment of SAS Viya order 923456, at the latest version of the Long Term 
Support (`lts`) cadence, with the contents going to file `./sas/923456_lts_depassets.tgz`: <br>

   ```
   go run main.go dep 923456 lts -p ./sas -f 923456_lts_depassets
   ```

   Sample output: <br>
     
   ```text
    OrderNumber: 923456
    AssetName: deploymentAssets
    AssetReqURL: https://api.sas.com/mysas/orders/923456/cadenceNames/lts/deploymentAssets
    AssetLocation: ./sas/923456_lts_depassets.tgz
    Cadence: Long Term Support 2020.0
    CadenceRelease: 20200808.1596943588306
   ```
   
* Get a renewal license to apply to the deployment of SAS Viya order 923456 from above, with the contents going to file 
`./sas/923456_lts_2020.0_license_ren1.jwt`: <br>
   
    ```go
      go run main.go lic 923456 lts 2020.0 -p ./sas -f 923456_lts_2020.0_license_ren1
    ```
   
   Sample output: <br>
     
   ```text
    OrderNumber: 923456
    AssetName: license
    AssetReqURL: https://api.sas.com/mysas/orders/923456/cadenceNames/lts/cadenceVersions/2020.0/license
    AssetLocation: ./sas/923456_lts_2020.0_license_ren1.jwt
    Cadence: Long Term Support 2020.0
    CadenceRelease:
   ```
  
* Get certificates for SAS Viya order 923457: <br>
   
   ```
   go run main.go cer 923457 -p ./sas -f 923457_certs -o json
   ``` 
   
   Sample output: <br>
     
   ```
   {
        "orderNumber": "923457",
        "assetName": "certificates",
        "assetReqURL": "https://api.sas.com/mysas/orders/923457/certificates",
        "assetLocation": "./sas/923457_certs.zip",
        "cadence": "",
        "cadenceRelease": ""
   }
   ```

## Contributing
We welcome your contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details 
on how to submit contributions to this project. 

## License
This project is licensed under the [Apache 2.0 License](LICENSE).

## Additional Resources
* [SAS API Portal](https://apiportal.sas.com/docs/mysasprod/1/overview)
* [SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html)
* [SAS Viya Operations Guide](https://documentation.sas.com/?softwareId=mysas&softwareVersion=prod&docsetId=itopswlcm&docsetTarget=home.htm)
