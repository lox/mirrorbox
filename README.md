Mirrorbox
=========

A local http server that servers the ubuntu mirrors.txt file with crappy mirrors removed.

Development
-----------

Requires geoip installed.

```bash
go get
go get github.com/mitchellh/gox
gox -build-toolchain
gox -osarch="darwin/amd64" -osarch="linux/amd64" -output="bin/{{.Dir}}_{{.OS}}_{{.Arch}}" ../mirrorbox/...
```