# GCU multi-device DRA e2e

Allocate two physical GCUs in one claim and verify both `/dev/gcuN` nodes
appear in the container.

The sample pins UUIDs `TR8C27121203` / `TR8C27150800` (`gcu-5` / `gcu-6` on the
lab node) to stay clear of `vllm-serve` and the single-card `gcu-basic` demo.
Update the CEL list if those cards are busy.

DRA names are `gcu-{char-minor}` and match `/dev/gcuN`. Prefer pinning by
UUID in CEL; the advertised UUID and injected node stay aligned via
`efmlDeviceGetMinorNumber`.

```fish
kubectl apply -f demo/examples/gcu-multi/gcu-multi.yaml
kubectl -n test-dra wait --for=condition=Ready pod/gcu-multi-pod --timeout=120s
kubectl -n test-dra exec gcu-multi-pod -- ls -la /dev/gcu* /dev/gcuctl
kubectl -n test-dra delete -f demo/examples/gcu-multi/gcu-multi.yaml
```
