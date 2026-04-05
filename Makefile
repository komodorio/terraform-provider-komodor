TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=github.com
TF_HOSTNAME=registry.terraform.io
NAMESPACE=komodorio
NAME=komodor
BINARY=terraform-provider-${NAME}
VERSION=2.4.0
OS_ARCH?=darwin_amd64

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --clean --snapshot --skip=publish,sign

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mkdir -p ~/.terraform.d/plugins/${TF_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${TF_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	rm ${BINARY}

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   

generate-docs:
	GO111MODULE=on GOFLAGS=-buildvcs=false go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name komodor

fmt:
	gofmt -w .
	terraform fmt -recursive examples/

lint:
	docker run --rm -v $(PWD):/workspace -w /workspace golangci/golangci-lint:v2.11.4 golangci-lint run --timeout=5m

check:
	bash scripts/check-local.sh
