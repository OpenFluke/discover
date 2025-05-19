# D.I.S.C.O.V.E.R.

Distributed Intelligent Scene Coordination & Virtual Entity Reconnaissance for Go.

## Install

```sh
go get github.com/OpenFluke/discover
```

## Usage

```
import "github.com/OpenFluke/discover"

func main() {
    cfg := discover.Config{ /* ... */ }
    d := discover.NewDiscover(cfg)
    d.ScanAll()
    d.PrintSummary()
}
```
