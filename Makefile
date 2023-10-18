.PHONY: all
all: help

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: run
run: ## Run controller locally.
	go run main.go --kubeconfig $(HOME)/.kube/config -v 4

.PHONY: apply-deployment
apply-deployment: ## Apply deployment.
	kubectl apply -f manifests/deployment.yaml

check-deployment: ## Check deployment.
	kubectl get deployment nginx-deployment -o yaml | grep -A 4 annotations:

check-event: ## Check event.
	kubectl events --for deployment/nginx-deployment -n default

.PHONY: launch-kind
launch-kind: ## Launch kind cluster.
	kind create cluster --name sample-controller

.PHONY: stop-kind
stop-kind: ## Stop kind cluster.
	kind delete cluster --name sample-controller
