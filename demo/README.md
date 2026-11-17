# Demo Examples

Workloads that exercise the Enflame GCU DRA driver:

| Example | What it does |
| --- | --- |
| [`gcu-basic`](examples/gcu-basic) | One GCU via ResourceClaim + Pod |
| [`gcu-multi`](examples/gcu-multi) | Two GCUs in one claim |

```fish
kubectl apply -f demo/examples/gcu-basic/gcu-basic.yaml
kubectl -n test-dra delete -f demo/examples/gcu-basic/gcu-basic.yaml
```
