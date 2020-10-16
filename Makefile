OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
ALL_PLATFORM = linux/amd64,linux/arm/v7,linux/arm64

# Image URL to use all building/pushing image targets
REGISTRY ?= scaleway
IMAGE ?= scaleway-operator
FULL_IMAGE ?= $(REGISTRY)/$(IMAGE)

IMAGE_TAG ?= $(shell git rev-parse HEAD)

DOCKER_CLI_EXPERIMENTAL ?= enabled

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"

ifdef TMPDIR
TMPDIR := $(realpath ${TMPDIR})
else
TMPDIR := /tmp
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: compile

generate-test-certs: CONFIGTXT := $(shell mktemp)
generate-test-certs: WEBHOOK_DIR := $(TMPDIR)/k8s-webhook-server
generate-test-certs: WEBHOOK_CERT_DIR := $(TMPDIR)/k8s-webhook-server/serving-certs
generate-test-certs:
	rm -rf $(WEBHOOK_DIR)
	mkdir -p $(WEBHOOK_CERT_DIR)

	@echo "[req]" > $(CONFIGTXT)
	@echo "distinguished_name = req_distinguished_name" >> $(CONFIGTXT)
	@echo "x509_extensions = v3_req" >> $(CONFIGTXT)
	@echo "req_extensions = req_ext" >> $(CONFIGTXT)
	@echo "[req_distinguished_name]" >> $(CONFIGTXT)
	@echo "[req_ext]" >> $(CONFIGTXT)
	@echo "subjectAltName = @alt_names" >> $(CONFIGTXT)
	@echo "[v3_req]" >> $(CONFIGTXT)
	@echo "subjectAltName = @alt_names" >> $(CONFIGTXT)
	@echo "[alt_names]" >> $(CONFIGTXT)
	@echo "DNS.1 = scaleway-operator-webhook-service.scaleway-operator-system.svc.cluster.local" >> $(CONFIGTXT)
	@echo "IP.1 = 127.0.0.1" >> $(CONFIGTXT)

	@echo "OpenSSL Config:"
	@cat $(CONFIGTXT)
	@echo

	openssl req -x509 -days 730 -out $(WEBHOOK_CERT_DIR)/tls.crt -keyout $(WEBHOOK_CERT_DIR)/tls.key -newkey rsa:4096 -subj "/CN=scaleway-operator-webhook-service.scaleway-operator-system" -config $(CONFIGTXT) -nodes

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out
	git diff --exit-code

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${FULL_IMAGE}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

compile:
	go build -v -o scaleway-operator main.go

docker-build:
	@echo "Building scaleway-operator for ${ARCH}"
	docker build . --platform=linux/$(ARCH) -f Dockerfile -t ${FULL_IMAGE}:${IMAGE_TAG}-$(ARCH)

docker-buildx-all:
	@echo "Making release for tag $(IMAGE_TAG)"
	docker buildx build --platform=$(ALL_PLATFORM) --push -t $(FULL_IMAGE):$(IMAGE_TAG) .

release: docker-buildx-all

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
