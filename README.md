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
    port := 80
    s := gonpm.NewProxy(port, storage.NewFS("/tmp/gonpm"))
    err := s.Listen(ctx)
    if err != nil {
        panic(err)
    }
}
```

## Deployment
```sh
make && make pushdev|pushstage|púhprod
```

