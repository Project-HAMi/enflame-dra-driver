# EFML

Vendored Enflame Management Library (`libefml`) headers for cgo builds.

## Headers

`include/efml/` is committed (from driver/EFML `1.9.2401`).

## Shared library (not committed)

Copy host libs before a linux cgo / container build:

```fish
mkdir -p third_party/efml/lib
scp enflame:/usr/lib/libefml.so* enflame:/usr/lib/libefml_static.a third_party/efml/lib/
```

Runtime nodes still need the Enflame driver package that provides `/usr/lib/libefml.so`.
