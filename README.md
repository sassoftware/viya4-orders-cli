# SAS Viya Orders CLI

SAS Viya Orders is a command-line interface (CLI) that calls the appropriate
[SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html)
endpoint to obtain the requested deployment assets for a specified SAS Viya
software order.

## Overview

You can use this CLI both as a tool and as an example of how to use Golang to
call the
[SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html).
You can also import the assetreqs and authn packages and use them in your own
Golang project.

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
  -n, --file-name string   name of the file where you want the downloaded order asset to be stored
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

- [Go](https://golang.org/) 1.13 or [Docker](https://www.docker.com/) is
  required.
- [git](https://git-scm.com/) version 2 or later is required.
- API credentials for the
  [SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html)
  are required. You can obtain them from the
  [SAS API Portal](https://apiportal.sas.com/get-started).

### Installation

#### Option 1 - Download a pre-built binary file.

Binaries for Windows, macOS, and Linux are available as downloads from
https://github.com/sassoftware/viya4-orders-cli/releases. Expand `Assets` under
the release of interest.

#### Option 2 - Build the project yourself.

To build the project, you have several options:

- Using [Make](https://www.gnu.org/software/make/): <br> Clone the project from
  the GitHub repo. Then, from the project root you can do the following:
  - Build a Windows executable (viya4-orders-cli_windows_amd64.exe) by running
    the following command:
    ```
    make win
    ```
  - Build a Linux executable (viya4-orders-cli_linux_amd64.exe) by running the
    following command:
    ```
    make linux
    ```
  - Build a macOS executable (viya4-orders-cli_darwin_amd64.exe) by running the
    following command:
    ```
    make darwin
    ```
  - Build all of the above executables by running the following command:
    ```
    make build
    ```
    Remove all of the above executables by running the following command:
    ```
    make clean
    ```
- Using [Docker](https://www.docker.com/): <br> You can build the project
  without cloning first:<br/><br/>
  - http:
    ```
    docker build github.com/sassoftware/viya4-orders-cli -t viya4-orders-cli
    ```
  - ssh:
    ```
    docker build git@github.com:sassoftware/viya4-orders-cli.git -t viya4-orders-cli
    ```
    Or you can clone the project from the GitHub repository. Then, from the
    project root, run the following command:
    ```
    docker build . -t viya4-orders-cli
    ```
- Using `go build`: <br> Clone the project from the GitHub repository. Then,
  from the project root, run the following command:
  ```
  go build -o {executable file name} main.go
  ```

## Getting Started

Take the following steps to start using SAS Viya Orders CLI:

1. If you do not yet have credentials for the
   [SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html),
   obtain them from the [SAS API Portal](https://apiportal.sas.com/get-started).

1. Record the `Key` and the `Secret` values that serve as your API credentials.
1. Base64 encode each value.
   > **NOTE:** When base64 encoding the credentials, take care not to encode
   > end-of-line characters into the result. <br/> Here is an example of the
   > correct way to encode from a Linux command prompt:<br/>
   > `echo -n {secret} | base64 --encode`<br/><br/> Here is an example of the
   > _incorrect_ way to encode from a Linux command prompt (the encoded result
   > will include `\n`):<br/> `echo {secret} | base64 --encode`
1. Add both credentials to your configuration file, or define them as
   environment variables:

   - Assign the encoded value of `Key` to `clientCredentialsId` (if using environment variables, use
    `CLIENTCREDENTIALSID`).
   - Assign the encoded value of `Secret` to `clientCredentialsSecret` (if using environment variables, use
    `CLIENTCREDENTIALSSECRET`).

1. Select CLI options. You can then specify them on the command line, pass them
   in as environment variables, or include them in a configuration file.

   SAS Viya Orders CLI options are applied in order of precedence, as follows:

   1. command-line specification
   1. environment variable specification
   1. config file specification
   1. default

1. If you want to use a configuration file, create it.

   The default location for the configuration file is `$HOME/.viya4-orders-cli`.
   You can save the file anywhere you want as long as you use the `--config` /
   `-c` option to inform the CLI of any non-default location.

   The config file must be in [YAML](https://yaml.org/) format, or else its file
   name must include a file extension that denotes another format. Supported
   formats are [JSON](https://www.json.org/),
   [TOML](https://github.com/toml-lang/toml), [YAML](https://yaml.org/),
   [HCL](https://github.com/hashicorp/hcl),
   [INI](<https://docs.microsoft.com/en-us/previous-versions/windows/desktop/ms717987(v=vs.85)>),
   [envfile](https://www.npmjs.com/package/envfile), or
   [Java properties](https://docs.oracle.com/javase/tutorial/essential/environment/properties.html)
   formats.

Here is a sample YAML configuration file that contains client credentials:

```
clientCredentialsId: 1a2B3c=
clientCredentialsSecret: 4D5e6F==
```

### Running

You have the following options for launching SAS Viya Orders CLI:

- Running the executable that you downloaded or built previously in the
  [Installation](#installation) section:

  ```
  {executable file name} [command] [args] [flags]
  ```

- Using Docker (this assumes that you have already executed the Docker build
  step described in the [Installation](#installation) section): <br>

  ```docker
  docker run viya4-orders-cli -v /my/local/path:/containerdir viya4-orders-cli [command] [args] [flags]
  ```

- Using `go run`: <br> If you have not done so already, clone the project from
  the GitHub repo. Then run the following command from the project root:

  ```
  go run main.go [command] [args] [flags]
  ```

### Examples

The examples in this section correspond to typical tasks that you might perform
using SAS Viya Orders CLI:

- Using a configuration file, `/c/Users/auser/vocli/.viya4-orders-cli.yaml`, to
  convey your API credentials, get deployment assets for SAS Viya order `923456`
  at the latest version of the Long Term Support (`lts`) cadence. Send the
  contents to file `/c/Users/auser/vocli/sasfiles/923456_lts_depassets.tgz`:
  <br>

  ```docker
  docker run -v /c/Users/auser/vocli:/sasstuff viya4-orders-cli deploymentAssets 923456 lts \
   --config /sasstuff/.viya4-orders-cli.yaml --file-path /sasstuff/sasfiles --file-name 923456_lts_depassets
  ```

  Sample output:

  ```text
  2020/10/02 19:16:30 Using config file: /sasstuff/.viya4-orders-cli.yaml
  OrderNumber: 923456
  AssetName: deploymentAssets
  AssetReqURL: https://api.sas.com/mysas/orders/923456/cadenceNames/lts/deploymentAssets
  AssetLocation: /sasstuff/sasfiles/923456_lts_depassets.tgz
  Cadence: Long Term Support 2020.0
  CadenceRelease: 20200808.1596943588306
  ```

- Get a renewal license for the deployment of SAS Viya order `923456` and send
  the contents to file
  `/auser/vocli/sasfiles/923456_lts_2020.0_license_ren1.jwt`:

  ```go
  go run main.go lic 923456 lts 2020.0 -p /auser/vocli/sasfiles -n 923456_lts_2020.0_license_ren1
  ```

  Sample output: <br>

  ```text
  OrderNumber: 923456
  AssetName: license
  AssetReqURL: https://api.sas.com/mysas/orders/923456/cadenceNames/lts/cadenceVersions/2020.0/license
  AssetLocation: /auser/vocli/sasfiles/923456_lts_2020.0_license_ren1.jwt
  Cadence: Long Term Support 2020.0
  CadenceRelease:
  ```

- Get certificates for SAS Viya order `923457` and send the contents to file
  `C:\Users\auser\vocli\sasfiles\923457_certs.zip`. Receive the output in JSON
  format:

  ```
  viya4-orders-cli_windows_amd64.exe cer 923457 -p C:\Users\auser\vocli\sasfiles -n 923457_certs -o json
  ```

  Sample output:

  ```
  {
      "orderNumber": "923457",
      "assetName": "certificates",
      "assetReqURL": "https://api.sas.com/mysas/orders/923457/certificates",
      "assetLocation": "C:\Users\auser\vocli\sasfiles\923457_certs.zip",
      "cadence": "",
      "cadenceRelease": ""
  }
  ```

## Contributing

We welcome your contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md)
for details on how to submit contributions to this project.

## License

This project is licensed under the [Apache 2.0 License](LICENSE).

## Additional Resources

- [SAS API Portal](https://apiportal.sas.com/docs/mysasprod/1/overview)
- [SAS Viya Orders API](https://developer.sas.com/guides/sas-viya-orders.html)
- [SAS Viya Operations Guide](https://documentation.sas.com/?softwareId=mysas&softwareVersion=prod&docsetId=itopswlcm&docsetTarget=home.htm)
