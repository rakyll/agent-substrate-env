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

# Multi-stage build for the ate-env service:
#
#   podman build --platform=linux/amd64 -t ate-env .
FROM golang:1.26 AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /ate-env ./cmd/ate-env

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ate-env /ate-env
ENTRYPOINT ["/ate-env"]
CMD ["serve"]
