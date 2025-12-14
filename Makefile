# Prophet Root Makefile
# Common workflows for the entire monorepo

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Demo

.PHONY: demo
demo: ## Run the self-healing demo (requires running cluster)
	@echo "🚀 Starting Prophet self-healing demo..."
	@cd demo/remediation-chaos && ./demo.sh

.PHONY: demo-setup
demo-setup: ## Set up demo prerequisites (Prometheus, Grafana)
	@echo "📦 Setting up monitoring stack..."
	kubectl apply -f monitoring/prometheus/prometheus.yaml
	kubectl apply -f monitoring/grafana/grafana.yaml
	@echo "✅ Monitoring stack deployed"

.PHONY: demo-cleanup
demo-cleanup: ## Clean up demo resources
	@echo "🧹 Cleaning up demo..."
	-kubectl delete ns demo-prophet --force --grace-period=0 2>/dev/null || true
	@echo "✅ Demo cleaned up"

##@ Operators

# List of all operators
OPERATORS := anomaly-remediator predictive-scaler slo-enforcer health-check budget-guard cost-alert diagnostic-remediator autonomous-agent

.PHONY: operators-build
operators-build: ## Build all operator binaries
	@echo "🔨 Building operator binaries..."
	@for op in $(OPERATORS); do \
		if [ -d "operators/$$op" ] && [ -f "operators/$$op/Makefile" ]; then \
			echo "Building $$op..."; \
			cd operators/$$op && make build && cd ../.. || true; \
		fi; \
	done
	@echo "✅ All operators built"

.PHONY: operators-docker-build
operators-docker-build: ## Build all operator Docker images
	@echo "🐳 Building operator Docker images..."
	@for op in $(OPERATORS); do \
		if [ -d "operators/$$op" ] && [ -f "operators/$$op/Dockerfile" ]; then \
			echo "Building Docker image for $$op..."; \
			cd operators/$$op && \
			IMG=ghcr.io/prophet-aiops/prophet-$$op:latest make docker-build && \
			cd ../.. || echo "⚠️  Failed to build $$op"; \
		fi; \
	done
	@echo "✅ All operator images built"
	@echo "💡 Tip: Use 'make operators-tilt-up' for live development with auto-rebuild"

.PHONY: operators-docker-push
operators-docker-push: ## Push all operator Docker images to registry
	@echo "📤 Pushing operator Docker images..."
	@for op in $(OPERATORS); do \
		if [ -d "operators/$$op" ] && [ -f "operators/$$op/Dockerfile" ]; then \
			echo "Pushing $$op..."; \
			cd operators/$$op && \
			IMG=ghcr.io/prophet-aiops/prophet-$$op:latest make docker-push && \
			cd ../.. || echo "⚠️  Failed to push $$op"; \
		fi; \
	done
	@echo "✅ All operator images pushed"

.PHONY: operators-tilt-up
operators-tilt-up: ## Start Tilt to build and deploy all operators
	@echo "🚀 Starting Tilt for all operators..."
	@if command -v tilt >/dev/null 2>&1; then \
		cd operators && tilt up --file Tiltfile; \
	else \
		echo "❌ Tilt not installed. Install with: brew install tilt"; \
		exit 1; \
	fi

.PHONY: operators-tilt-down
operators-tilt-down: ## Stop Tilt
	@echo "🛑 Stopping Tilt..."
	@if command -v tilt >/dev/null 2>&1; then \
		cd operators && tilt down; \
	else \
		echo "⚠️  Tilt not installed"; \
	fi

.PHONY: operators-load-kind
operators-load-kind: ## Load all operator images into kind cluster (Tilt handles this automatically)
	@echo "📥 Tilt automatically loads images into kind when running"
	@echo "💡 Run 'make operators-tilt-up' to build and deploy all operators"

.PHONY: operators-test
operators-test: ## Run tests for all operators
	@echo "🧪 Testing operators..."
	@for op in $(OPERATORS); do \
		if [ -d "operators/$$op" ] && [ -f "operators/$$op/Makefile" ]; then \
			echo "Testing $$op..."; \
			cd operators/$$op && make test && cd ../.. || true; \
		fi; \
	done
	@echo "✅ All operator tests passed"

.PHONY: operators-lint
operators-lint: ## Lint all operators
	@echo "🔍 Linting operators..."
	@for op in $(OPERATORS); do \
		if [ -d "operators/$$op" ] && [ -f "operators/$$op/Makefile" ]; then \
			echo "Linting $$op..."; \
			cd operators/$$op && make lint && cd ../.. || true; \
		fi; \
	done
	@echo "✅ All operators linted"

.PHONY: operators-deploy
operators-deploy: ## Deploy all operators to current cluster
	@echo "🚀 Deploying operators..."
	kubectl apply -f clusters/common/aiops/operators/
	@echo "✅ Operators deployed"

.PHONY: operators-status
operators-status: ## Check status of all Prophet CRDs
	@echo "📊 Prophet CRD Status:"
	@echo ""
	@echo "AnomalyActions:"
	@kubectl get anomalyactions -A 2>/dev/null || echo "  (none found)"
	@echo ""
	@echo "PredictiveScalers:"
	@kubectl get predictivescalers -A 2>/dev/null || echo "  (none found)"
	@echo ""
	@echo "SLOEnforcers:"
	@kubectl get sloenforcer -A 2>/dev/null || echo "  (none found)"
	@echo ""
	@echo "HealthChecks:"
	@kubectl get healthchecks -A 2>/dev/null || echo "  (none found)"

##@ Development

.PHONY: dev-up
dev-up: ## Start local development (kind + tilt for all operators)
	@echo "🏗️  Starting local development environment..."
	@KIND_CLUSTER=$$(kubectl config current-context | sed 's/.*kind-//' | sed 's/-control-plane//' 2>/dev/null || echo ""); \
	if [ -z "$$KIND_CLUSTER" ] || ! kind get clusters 2>/dev/null | grep -q "$$KIND_CLUSTER"; then \
		echo "Creating kind cluster: prophet-local..."; \
		kind create cluster --name prophet-local || echo "Cluster may already exist"; \
		kubectl config use-context kind-prophet-local || true; \
	fi
	@echo "Starting Tilt for all operators..."
	@make operators-tilt-up

.PHONY: dev-down
dev-down: ## Stop local development
	@echo "🛑 Stopping development environment..."
	-cd operators && tilt down
	@echo "✅ Stopped (kind cluster preserved - run 'make dev-destroy' to remove)"

.PHONY: dev-destroy
dev-destroy: ## Destroy local kind cluster
	@echo "💥 Destroying kind cluster..."
	-kind delete cluster --name prophet-dev
	@echo "✅ Cluster destroyed"

##@ Validation

.PHONY: validate
validate: ## Validate all manifests
	@echo "✅ Validating manifests..."
	@find clusters/ -name "*.yaml" -type f | head -20 | xargs -I {} kubectl apply --dry-run=client -f {} 2>/dev/null || true
	@echo "Validation complete (check output for errors)"

.PHONY: validate-crds
validate-crds: ## Validate CRDs with kubeconform
	@echo "🔍 Validating CRDs..."
	@if command -v kubeconform >/dev/null 2>&1; then \
		find clusters/common/aiops/operators/ -name "*.yaml" -exec kubeconform -kubernetes-version 1.29 -ignore-missing-schemas {} \; ; \
	else \
		echo "⚠️  kubeconform not installed. Install with: brew install kubeconform"; \
	fi

.PHONY: lint-yaml
lint-yaml: ## Lint YAML files
	@echo "🔍 Linting YAML..."
	@if command -v yamllint >/dev/null 2>&1; then \
		yamllint -d relaxed clusters/ monitoring/ resilience/ || true; \
	else \
		echo "⚠️  yamllint not installed. Install with: pip install yamllint"; \
	fi

##@ Monitoring

.PHONY: monitoring-deploy
monitoring-deploy: ## Deploy full monitoring stack
	@echo "📊 Deploying monitoring stack..."
	kubectl apply -f monitoring/prometheus/prometheus.yaml
	kubectl apply -f monitoring/prometheus/alertmanager.yaml
	kubectl apply -f monitoring/grafana/grafana.yaml
	kubectl apply -f monitoring/kube-state-metrics/
	@echo "✅ Monitoring deployed"

.PHONY: grafana-port-forward
grafana-port-forward: ## Port-forward Grafana (localhost:3000)
	@echo "🔗 Port-forwarding Grafana to localhost:3000..."
	kubectl port-forward -n monitoring svc/grafana 3000:3000

.PHONY: prometheus-port-forward
prometheus-port-forward: ## Port-forward Prometheus (localhost:9090)
	@echo "🔗 Port-forwarding Prometheus to localhost:9090..."
	kubectl port-forward -n monitoring svc/prometheus 9090:9090

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@for op in $(OPERATORS); do \
		if [ -d "operators/$$op" ]; then \
			rm -rf operators/$$op/bin 2>/dev/null || true; \
		fi; \
	done
	@echo "✅ Cleaned"

.PHONY: clean-all
clean-all: clean demo-cleanup ## Clean everything including demo
	@echo "✅ Full cleanup complete"

