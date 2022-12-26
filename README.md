<a href="https://terraform.io">
    <img src=".github/tf.png" alt="Terraform logo" title="Terraform" align="left" height="50" />
</a>

# Terraform Provider for Komodor

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

![Terraform](https://rawgithub.com/hashicorp/terraform/master/website/source/assets/images/logo-hashicorp.svg)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18

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

## Testing the Provider

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
