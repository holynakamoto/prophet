# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/prophet-aiops/prophet-label-enforcer:latest
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false,allowDangerousTypes=true"

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

KUSTOMIZE_VERSION ?= v5.3.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0

.PHONY: all
all: build

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd:allowDangerousTypes=true webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="" paths="./..."

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: manifests generate fmt vet
	go test ./... -coverprofile cover.out

.PHONY: build
build: generate fmt vet
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet
	go run ./cmd/main.go

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

# Helm targets
.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	helm lint helm/label-enforcer

.PHONY: helm-package
helm-package: ## Package the Helm chart
	helm package helm/label-enforcer

.PHONY: helm-template
helm-template: ## Show the Helm templates
	helm template label-enforcer helm/label-enforcer

.PHONY: helm-install
helm-install: ## Install the Helm chart
	helm upgrade --install label-enforcer helm/label-enforcer

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall the Helm chart
	helm uninstall label-enforcer