# Turn off proxy before build. Otherwise `go build` fails.
go env -w GO111MODULE=off