# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Integration test target deployment. Override on the command line to match
# your cluster, e.g.:
#
#   make test-integration ATE_TEST_ATESPACE=ax ATE_TEST_TEMPLATE=ax-harness-template
ATE_TEST_ATEAPI    ?= localhost:8443
ATE_TEST_ATESPACE  ?= default
ATE_TEST_TEMPLATE  ?= default-env-template

.PHONY: build test test-integration port-forward-ateapi vet

build:
	go build ./...

# Unit tests only; integration tests are excluded by the "integration" build tag.
test:
	go test ./...

# End-to-end tests against a real Agent Substrate deployment. Requires the
# control API to be reachable at ATE_TEST_ATEAPI — run "make port-forward-ateapi"
# in another terminal first (or point ATE_TEST_ATEAPI at an in-cluster address).
integration:
	ATE_TEST_ATEAPI=$(ATE_TEST_ATEAPI) \
	ATE_TEST_ATESPACE=$(ATE_TEST_ATESPACE) \
	ATE_TEST_TEMPLATE=$(ATE_TEST_TEMPLATE) \
	go test -tags=integration -v ./cmd/ate-env/

# Forward the Agent Substrate control API to localhost:8443 (blocks).
port-forward-ateapi:
	kubectl port-forward -n ate-system svc/api 8443:443
