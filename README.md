# GoNPM
A customized npm proxy written in Go


## Development
- Go version >= 1.10
- `cd cmd; go run .`

## Use as a library
```go
import "github.com/truongminh/gonpm"

func main() {
    port := 80
    s := gonpm.NewProxy(port)
    err := s.Listen(ctx)
    if err != nil {
        panic(err)
    }
}
```

