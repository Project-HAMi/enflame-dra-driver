# Enflame GCU DRA Driver

Kubernetes [Dynamic Resource Allocation (DRA)](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
driver for Enflame GCU devices.

- Module: `github.com/Project-HAMi/enflame-dra-driver`
- Default driver name: `gcu.enflame.com`
- CDI specs: `/var/run/cdi` (`k8s.gcu.enflame.com/gcu`)

## Binaries

| Binary | Role |
| --- | --- |
| `enflame-dra-kubeletplugin` | Node kubelet DRA plugin |
| `enflame-dra-webhook` | Optional validating webhook (opaque config; stage 6) |
| `gcu-cdi-gen` | Offline CDI spec generator for host comparison |

## Build

```bash
make binaries
make test
make lint
```

Helm chart: `deployments/helm/enflame-dra-driver`.

## Deploy (sketch)

Plugin DaemonSet mounts the host root at `/host` (`HOST_ROOT=/host`) so CDI discovery can see firmware / `efsmi` / `libefml`, plus `/dev` and host `/usr/lib` for the plugin process. Default driver name is `gcu.enflame.com`; webhook stays off until `GcuConfig` exists.

```fish
# on a linux/amd64 GCU node with Go + libefml
set -x GOPROXY https://goproxy.cn,direct
set -x CGO_ENABLED 1
go build -o enflame-dra-kubeletplugin ./cmd/enflame-dra-kubeletplugin
go build -o enflame-dra-webhook ./cmd/enflame-dra-webhook
docker build -f deployments/container/Dockerfile.runtime \
  --build-arg BASE_IMAGE=ubuntu:24.04 \
  -t enflame-dra-driver:dev .
# containerd (kubelet) needs the image in the k8s.io namespace
docker save enflame-dra-driver:dev | ctr -n k8s.io images import -

helm upgrade --install enflame-dra-driver deployments/helm/enflame-dra-driver \
  -n kube-system \
  --set image.repository=enflame-dra-driver \
  --set image.tag=dev \
  --set image.pullPolicy=Never \
  --set 'kubeletPlugin.tolerations[0].operator=Exists'
```

## Status

- Stage 0: CDI manager — done
- Stage 1: scaffold rename / mock profile removal — done
- Stage 2: EFML device discovery (`pkg/gcu` via libefml cgo; sysfs only for NUMA/PCIe topology) — done
- Stage 3: ResourceSlice attributes aligned with HAMi nvidia-dra-driver (`type`/`uuid`/`productName`/`brand`/`pcieBusID`/`driverVersion` + standard `pcieRoot`/`numaNode`; GCU extras `serialNumber`/`firmwareVersion`; capacity `memory`) — done
- Stage 4: Prepare/Unprepare injects real `/dev/gcuN` via CDI; startup cleans orphaned claim specs — done
- Stage 5: Device health monitoring via sysfs `device_health`/`device_state`, reported through KEP-4680 `WatchHealthStatus` — done
- Stage 7: Deployed on enflame (`HOST_ROOT=/host`, local `enflame-dra-driver:dev`, driver `gcu.enflame.com`) — done
- Stage 8: e2e Claim+Pod — `demo/examples/gcu-basic` (1 card) and `demo/examples/gcu-multi` (2 cards); `/dev/gcuN` + `/dev/gcuctl` injected — done
- Deferred: `GcuConfig` / webhook
- Fix: DRA name / CDI path / sysfs health use `efmlDeviceGetMinorNumber` (`/dev/gcu{minor}`); EFML index is kept only as a logical field (may differ from minor). Verified on enflame: claim UUIDs `TR8C27121203`/`TR8C27150800` → `gcu-3`/`gcu-4` → `/dev/gcu3`/`/dev/gcu4`, container `efsmi` matches.
