<a href="https://terraform.io">
    <img src="tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

# Terraform Provider for Komodor

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.3.x
- [Go](https://golang.org/doc/install) >= 1.18 ((to build the provider plugin)

## Installation

Add the following to your terraform configuration

```tf
terraform {
  required_providers {
    komodor = {
      source  = "komodorio/komodor"
      version = "~> 1.0.4"
    }
  }
}
```

## How to use

First, you need a [Komodor](https://komodor.com/) account.

Once you have the account, you should create an API key. Go to the **API Keys** tab in the [User Settings Page](https://app.komodor.com/settings/api-keys). Click on **Generate Key** button and generate the key.

Configure the terraform provider like so

```tf
variable "komodor_api_key" {
  type = string
}

provider "komodor" {
  api_key = var.komodor_api_key
}
```

To see examples of how to use this provider, check out the `examples` directory in the source code [here](/examples).

## Developing The Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.18+ is _required_). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

Clone the repository:

```sh
mkdir -p $GOPATH/src/github.com/terraform-providers; cd "$_"
git clone https://github.com/komodor/terraform-provider-komodor.git
```

Change to the clone directory and run make tools to install the dependent tooling needed to test and build the provider.

To compile the provider, run make build. This will build the provider and put the provider binary in the $GOPATH/bin directory.

```sh
cd $GOPATH/src/github.com/komodor/terraform-provider-komodor
make tools
make build
$GOPATH/bin/terraform-provider-komodor
```
