# EFML

Vendored Enflame Management Library (`libefml`) for linux/amd64 cgo builds and
container image releases.

## Layout

| Path | Role |
| --- | --- |
| `include/efml/` | Headers (from EFML `1.9.2401`) |
| `lib/libefml.so*` | Shared library used by `#cgo LDFLAGS: -lefml` |
| `lib/libefml_static.a` | Static archive from the same package |

Runtime nodes still need the Enflame driver package that provides host
`/usr/lib/libefml.so` (and device nodes). The copies here are for building the
DRA plugin binary and release images on CI / developer machines without a GCU
host.

## Refreshing from a driver node

```fish
mkdir -p third_party/efml/lib
scp enflame:/usr/lib/libefml.so.1.9.2401 enflame:/usr/lib/libefml_static.a third_party/efml/lib/
cd third_party/efml/lib
ln -sf libefml.so.1.9.2401 libefml.so.1
ln -sf libefml.so.1 libefml.so
```

Keep the soname symlinks; avoid committing three identical `.so` blobs.
