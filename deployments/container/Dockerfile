# Copyright 2023 The Kubernetes Authors.
# Copyright The Project-HAMi Authors.
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

ARG GO_VERSION=undefined
ARG BASE_IMAGE=undefined
FROM golang:${GO_VERSION} AS build

# Default to a China-reachable module proxy; override at build time if needed.
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /build
COPY . .

# libefml is required for cgo. Place shared libs under third_party/efml/lib
# (see third_party/efml/README.md) before building the image.
RUN mkdir -p /usr/lib /out/lib && \
    if ls third_party/efml/lib/libefml.so* >/dev/null 2>&1; then \
      cp -a third_party/efml/lib/libefml.so* /usr/lib/ /out/lib/; \
    fi

RUN mkdir /artifacts
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux make PREFIX=/artifacts cmds

FROM ${BASE_IMAGE}

LABEL io.k8s.display-name="Enflame GCU Driver for Dynamic Resource Allocation (DRA)"
LABEL name="Enflame GCU Driver for Dynamic Resource Allocation (DRA)"
LABEL vendor="enflame.com"
ARG VERSION="N/A"
LABEL version=${VERSION}
ARG RELEASE="N/A"
LABEL release="N/A"
LABEL summary="Enflame GCU DRA driver for Kubernetes"
LABEL description="See summary"

COPY --from=build /out/lib/ /usr/lib/
COPY --from=build /artifacts/enflame-dra-kubeletplugin /usr/bin/enflame-dra-kubeletplugin
COPY --from=build /artifacts/enflame-dra-webhook       /usr/bin/enflame-dra-webhook
