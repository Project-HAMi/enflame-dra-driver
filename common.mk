# Copyright 2022 The Kubernetes Authors.
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

# Matches `toolchain` in go.mod
GO_VERSION ?= 1.26.8

DRIVER_NAME := enflame-dra-driver
MODULE := github.com/Project-HAMi/$(DRIVER_NAME)

VERSION  ?=
vVERSION := v$(VERSION:v%=%)

VENDOR := enflame.com
APIS := $(CURDIR)/internal/api/checkpoint $(CURDIR)/internal/api/checkpoint/v1

ifeq ($(IMAGE_NAME),)
REGISTRY ?= dynamia-ai-registry.cn-hangzhou.cr.aliyuncs.com/artifacts/dynamia-ai
IMAGE_NAME = $(REGISTRY)/$(DRIVER_NAME)
endif
