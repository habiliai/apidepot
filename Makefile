SHELL := $(shell which sh)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)
PROTOC_DIR := bin/protoc-$(GOOS)-$(GOARCH)
PROTOC := bin/protoc
APIDEPOT_CTL := bin/apidepotctl
APIDEPOT := bin/apidepot
PB_FILES := pkg/internal/proto/apidepot.pb.go pkg/internal/proto/apidepot_grpc.pb.go

.: all

.PHONY: all
all: $(APIDEPOT_CTL) $(APIDEPOT)

$(APIDEPOT): pb
	CGO_ENABLED=0 go build -o $@ cmd/apidepot/main.go

$(APIDEPOT_CTL)-%: pb
	@CGO_ENABLED=0 GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(lastword $(subst -, ,$@)) go build -o $@ cmd/apidepotctl/main.go
	@echo "done $@"

$(APIDEPOT_CTL): $(APIDEPOT_CTL)-$(GOOS)-$(GOARCH)
	ln -sf $(APIDEPOT_CTL)-$(GOOS)-$(GOARCH) $(APIDEPOT_CTL)

bin/protoc-linux-amd64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-linux-x86_64.zip"

bin/protoc-linux-arm64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-linux-aarch_64.zip"

bin/protoc-darwin-amd64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-osx-x86_64.zip"

bin/protoc-darwin-arm64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-osx-aarch_64.zip"

$(PROTOC_DIR): bin/protoc-$(GOOS)-$(GOARCH).zip
	@unzip -o $< -d $(PROTOC_DIR)
	@echo "done $@"

$(PROTOC): $(PROTOC_DIR)
	chmod 755 $(PROTOC_DIR)/bin/protoc
	ln -sf protoc-$(GOOS)-$(GOARCH)/bin/protoc $(PROTOC)
	touch $(PROTOC)

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34

PROTOC_GEN_GO_GRPC := $(GOPATH)/bin/protoc-gen-go-grpc
$(PROTOC_GEN_GO_GRPC): $(PROTOC_GEN_GO)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4

%.pb.go: %.proto $(PROTOC) $(PROTOC_GEN_GO)
	@export PATH="$(shell go env GOPATH)/bin:$(PATH)"
	$(PROTOC) --go_out=. --go_opt=paths=source_relative $<

%_grpc.pb.go: %.proto $(PROTOC) $(PROTOC_GEN_GO)
	@export PATH="$(shell go env GOPATH)/bin:$(PATH)"
	$(PROTOC) --go-grpc_out=. --go-grpc_opt=paths=source_relative $<

.PHONY: pb
pb: $(PB_FILES)

.PHONY: test
test: pb
	CI=true go test -timeout 15m -p 1 ./...

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf $(PB_FILES)
	rm -rf $(PROTOC_DIR)
	rm -f $(PROTOC)
	@echo "cleared"

GOLANGCI_LINT := ./bin/golangci-lint
$(GOLANGCI_LINT):
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.62.2
	@chmod +x ./bin/golangci-lint
	@echo "golangci-lint installed"

.PHONY: lint
lint: $(GOLANGCI_LINT) $(PB_FILES)
	$(GOLANGCI_LINT) run

KIND := $(GOPATH)/bin/kind
$(KIND):
	go install sigs.k8s.io/kind@v0.20.0

setup: $(KIND)
	$(SHELL) testdata/kind/download-docker-images.sh
	kind create cluster --config testdata/kind/kind.yaml
	kubectl wait --context kind-apidepot --for=condition=ready node --selector=type=default --timeout=60s

	kind load docker-image $(shell cat testdata/kind/docker_images.txt) --name apidepot --verbosity 1

	bash -e testdata/kind/yamls/metallb/metallb.sh
	bash -e testdata/kind/yamls/traefik/traefik.sh
	kubectl apply --context kind-apidepot -f "testdata/kind/yamls/*"
	kubectl rollout --context kind-apidepot -n kube-system restart deploy/coredns
	kubectl wait --context kind-apidepot --namespace default --for=condition=ready --timeout=300s pod -l group=apidepot

teardown: $(KIND)
	$(KIND) delete cluster --name apidepot