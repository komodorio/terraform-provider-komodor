#!/usr/bin/env bash
# Validates all Terraform example configs against the locally-built provider binary.
# Requires: go, unzip, curl
set -euo pipefail

TF_VERSION="1.9.0"
BINARY_DIR="$(pwd)"

echo "==> Installing Terraform ${TF_VERSION}..."
if ! command -v terraform &>/dev/null; then
  # Detect OS and architecture
  case "$(uname -s)" in
    Linux)  TF_OS="linux" ;;
    Darwin) TF_OS="darwin" ;;
    *)      echo "Unsupported OS: $(uname -s)"; exit 1 ;;
  esac

  case "$(uname -m)" in
    x86_64)          TF_ARCH="amd64" ;;
    arm64 | aarch64) TF_ARCH="arm64" ;;
    *)               echo "Unsupported arch: $(uname -m)"; exit 1 ;;
  esac

  TF_ZIP="/tmp/terraform_${TF_VERSION}_${TF_OS}_${TF_ARCH}.zip"
  curl -fsSL "https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_${TF_OS}_${TF_ARCH}.zip" -o "${TF_ZIP}"

  # Install to a user-writable location if /usr/local/bin requires root
  TF_INSTALL_DIR="/usr/local/bin"
  if [ ! -w "${TF_INSTALL_DIR}" ]; then
    TF_INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "${TF_INSTALL_DIR}"
    export PATH="${TF_INSTALL_DIR}:${PATH}"
  fi
  unzip -q "${TF_ZIP}" -d "${TF_INSTALL_DIR}"
fi
terraform version

echo "==> Building provider binary..."
GO111MODULE=on go build -o "${BINARY_DIR}/terraform-provider-komodor" .

echo "==> Writing dev override config..."
cat > /tmp/dev.terraformrc <<EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/komodorio/komodor" = "${BINARY_DIR}"
  }
  direct {}
}
EOF
export TF_CLI_CONFIG_FILE=/tmp/dev.terraformrc

FAILED=0

validate_dir() {
  local dir="$1"
  local injected="${dir}/_validate_required_providers.tf"
  echo "--> Validating ${dir}..."

  # Inject a required_providers block so terraform resolves the provider via the
  # dev override (komodorio/komodor) rather than defaulting to hashicorp/komodor.
  # Skip injection if any .tf file in the directory already declares required_providers.
  # This file is removed immediately after validation regardless of outcome.
  local injected_this_run=0
  if ! grep -rl "required_providers" "${dir}"*.tf &>/dev/null; then
    cat > "${injected}" <<'EOF'
terraform {
  required_providers {
    komodor = {
      source = "komodorio/komodor"
    }
  }
}
EOF
    injected_this_run=1
  fi

  local result=0
  # Suppress the expected dev-override warning; surface everything else.
  terraform -chdir="${dir}" validate -no-color 2>&1 \
    | grep -v "Provider development overrides are in effect" \
    | grep -v "komodorio/komodor in" \
    | grep -v "The behavior may therefore not match" \
    | grep -v "applying changes may cause" \
    || result=1

  [ "${injected_this_run}" -eq 1 ] && rm -f "${injected}"

  if [ "${result}" -ne 0 ]; then
    echo "FAILED (validate): ${dir}"
    FAILED=1
  fi
}

for dir in examples/resources/*/; do
  validate_dir "${dir}"
done

for dir in examples/data-sources/*/; do
  validate_dir "${dir}"
done

if [ "${FAILED}" -ne 0 ]; then
  echo ""
  echo "ERROR: One or more example directories failed validation."
  echo "Fix the examples to match the current provider schema."
  exit 1
fi

echo "All examples validated successfully."
