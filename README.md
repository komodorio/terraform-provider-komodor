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
      version = "~> 1.0.6"
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

### Using EU Region

To use the Komodor EU region, set the `api_url` parameter:

```tf
provider "komodor" {
  api_key = var.komodor_api_key
  api_url = "https://api.eu.komodor.com"
}
```

Alternatively, you can set the `KOMODOR_API_URL` environment variable:

```sh
export KOMODOR_API_URL="https://api.eu.komodor.com"
```

By default, the provider uses `https://api.komodor.com` (US region) if `api_url` is not specified.

To see examples of how to use this provider, check out the `examples` directory in the source code [here](/examples).

## Developing The Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.18+ is _required_). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

Clone the repository:

```sh
mkdir -p $GOPATH/src/github.com/terraform-providers; cd "$_"
git clone https://github.com/komodorio/terraform-provider-komodor.git
```

Change to the clone directory and run make to install the dependent tooling needed to test and build the provider.

To compile the provider, run make build. This will build the provider and put the provider binary in the $GOPATH/bin directory.

```sh
cd $GOPATH/src/github.com/komodor/terraform-provider-komodor
make 
$GOPATH/bin/terraform-provider-komodor
```

To build the provider for a specific OS and architecture, run make with the OS_ARCH variable set to the desired target. For example, to build the provider for macOS ARM CPU, run make with OS_ARCH=darwin_arm64.

```sh
make OS_ARCH=darwin_arm64
```

### Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile the provider binary |
| `make install` | Build and install the provider locally for manual testing |
| `make fmt` | Auto-format Go source files and Terraform example files |
| `make lint` | Run `golangci-lint` |
| `make generate-docs` | Regenerate `docs/` from templates and schema |
| `make check` | Run all CI checks locally and print a pass/fail summary |

### Running All Checks Locally

```sh
make check
```

This runs every CI check in sequence and prints a pass/fail summary:

- **fmt check** — detects unformatted Go files
- **mod tidy** — ensures `go.mod` / `go.sum` are in sync
- **vet** — runs `go vet`
- **lint** — runs `golangci-lint`
- **unit tests** — `go test -race ./komodor/...`
- **build** — verifies the provider compiles
- **docs check** — regenerates docs and fails if `docs/` is out of date
- **examples fmt** — checks that Terraform example files are formatted
- **examples validate** — validates all example configs against the provider schema
- **goreleaser check** — validates the release config
- **acc test coverage** — ensures every resource has an acceptance test file
- **acceptance tests** — skipped unless `KOMODOR_API_KEY` is set (see below)

### Acceptance Tests

Acceptance tests create and destroy real resources in Komodor. To run them, provide your API key:

```sh
KOMODOR_API_KEY=<your-api-key> make check
```

Or run them directly:

```sh
KOMODOR_API_KEY=<your-api-key> TF_ACC=1 go test -v -run TestAcc ./komodor/... -timeout 60m
```

All test resources are created with the `tf-acc-` name prefix. Any resources left over from a crashed run are automatically cleaned up at the start of the next test run.

### Adding a New Resource

1. Implement the resource in `komodor/resource_komodor_<name>.go`
2. Register it in `provider.go` under `ResourcesMap`
3. Add an acceptance test file `komodor/resource_komodor_<name>_acc_test.go` with:
   ```go
   func init() { registerAccTest("komodor_<name>") }
   ```
4. Add example configs under `examples/resources/komodor_<name>/`
5. Add a template under `templates/resources/<name>.md.tmpl` (or let `make generate-docs` auto-generate one)
6. Run `make generate-docs` and commit the updated `docs/`

> **CI enforces test coverage**: the `acc test coverage` step fails if any resource in `ResourcesMap` has no corresponding acceptance test. Adding a resource without a test file will block CI.

### Updating Docs

Files under `docs/` are fully generated — **do not edit them directly**. Instead:

- Edit `templates/resources/<name>.md.tmpl` for resource-specific content and examples
- Add or update example `.tf` files under `examples/`
- Run `make generate-docs` to regenerate `docs/`
- Commit both the template/example changes and the regenerated `docs/`
