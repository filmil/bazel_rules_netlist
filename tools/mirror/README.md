<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mirror

`manifest.json` lists the files this repository mirrors as release
assets, and where each comes from. `bazel run //tools:mirror` checks
every upstream file against it; the operator uploads with
`--dry_run=false --built_dir=DIR` (see `cmd/main.go`).

## SKiDL wheels

PyPI has only sdists of skidl 3.0.0, kinet2pcb 1.1.4 and hierplace 1.1.0,
which are pure Python. The SKiDL toolchain (issue #9) installs wheels
only, so their wheels are built once and mirrored as the release
`mirror-skidl-3.0.0`. The manifest names each sdist and its sha256.

To build them, in the pinned Python image, into `/tmp/wheels`:

```sh
mkdir -p /tmp/wheels
podman run --rm -v /tmp/wheels:/w \
  -e SOURCE_DATE_EPOCH=315532800 -e PIP_DISABLE_PIP_VERSION_CHECK=1 \
  docker.io/library/python@sha256:aa943ce411423c7582288f18707ff66eddcbaeb510502d747a72aeb34f30d159 \
  pip wheel --quiet --no-deps --no-binary :all: -w /w \
    'skidl @ https://files.pythonhosted.org/packages/source/s/skidl/skidl-3.0.0.tar.gz#sha256=265f2b188a87c4e7190a46e08ac76e8e010b00232aef5093f5d801e062885b0e' \
    'kinet2pcb @ https://files.pythonhosted.org/packages/source/k/kinet2pcb/kinet2pcb-1.1.4.tar.gz#sha256=9374a09ba58f64895b5e47c972b20315d624e06eec48925df26a308e2746f661' \
    'hierplace @ https://files.pythonhosted.org/packages/source/h/hierplace/hierplace-1.1.0.tar.gz#sha256=1386717bb9e6759d9b319013d6edd0f568b717f834e29d9d804ab5cc4c0d0ad0'
bazel run //tools:mirror -- --built_dir=/tmp/wheels
```

pip checks each sdist against the `#sha256=` of its URL.
`SOURCE_DATE_EPOCH` fixes the times in the wheels, so two builds give the
same bytes; the last command checks them against the manifest.
