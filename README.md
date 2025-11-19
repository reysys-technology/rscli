```shell
go mod verify ;
go mod tidy ;
gofmt -s -e -w . ;
go fix ./... ;
go vet ./...
```