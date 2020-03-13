# gonpm
A modularized npm cache proxy written in Go

## Development
- Go version >= 1.10
- Start server: `cd cmd; PORT=8999 NPM_CACHE=mem go run .`
- Test npm install: `cd cmd/test; sh install.sh`

## Use as a library
```go
import "github.com/truongminh/gonpm"
import "github.com/truongminh/gonpm/storage"

func main() {
    port := 8080
    cacheURI := "fs:///tmp?limit=256GB"
    store, err := storage.Open(cacheURI)
    if err != nil {
        panic(err)
    }
    s := gonpm.NewProxy(port, store)
    err := s.Listen(ctx)
    if err != nil {
        panic(err)
    }
}
```

## Deployment
```sh
make && make pushdev|pushstage|pushprod
```

