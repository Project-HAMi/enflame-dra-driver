# GCU basic DRA e2e

Allocate one physical GCU through `gcu.enflame.com` and verify CDI injection
(`/dev/gcuN`, `/dev/gcuctl`, `efsmi`).

The sample claim pins UUID `TR6W18220404` (`gcu-0` on the lab node) so it does
not collide with the long-running `vllm-serve` workload on other cards. Change
the CEL expression if that UUID is busy.

```fish
kubectl apply -f demo/examples/gcu-basic/gcu-basic.yaml
kubectl -n test-dra wait --for=condition=Ready pod/gcu-basic-pod --timeout=120s
kubectl -n test-dra exec gcu-basic-pod -- efsmi -L
kubectl -n test-dra delete -f demo/examples/gcu-basic/gcu-basic.yaml
```
