# Contributing

Ideas and fixes are welcome. Open an issue first if the change is large.

```sh
gofmt -l .
go vet ./...
go test ./...
```

All three must be clean. The packages are in [what is where](docs/3-what-is-where.md); each test sits
beside the code it proves, and every change in behaviour comes with one. Most changes can be tried
with `apptest`; say in the pull request whether yours ran on a real tablet.

Never add anything taken from the tablet: its binaries, or tables derived from them, are not ours to
share.
