## Golang SDK for Slicer

SDK for [SlicerVM.com](https://slicervm.com)

### Installation

```bash
go get github.com/slicervm/sdk@latest
```

### Example usage

Create a new slicer config:

```bash
slicer new api \
    --count=0 \
    --graceful-shutdown=false \
    --ram 4 \
    --cpu 2 > api.yaml
```

Create a VM (node) in a host group with the default RAM/CPU settings as defined in the host group.

```go
package main

import (
    "fmt"
    "os"
    sdk "github.com/slicervm/sdk"
)

func main() {
    // Typically you'd load these from environment variables
    baseURL := os.Getenv("SLICER_URL")      // API base URL
    token := os.Getenv("SLICER_TOKEN")      // Your API token
    userAgent := "my-microvm-client/1.0"
    hostGroup := "api"                       // Existing host group name

    client := sdk.NewSlicerClient(baseURL, token, userAgent, nil /* or &http.Client{} */)

    createReq := sdk.SlicerCreateNodeRequest{
        RAMGB:      4,
        CPUs:       2,
        Userdata: `#!/bin/bash
echo 'Bootstrapping...'
ping -c3 google.com

sudo reboot
`,
        SSHKeys: []string{"ssh-rsa AAAA..."}, // Optional: inject public SSH keys
        ImportUser: "alexellis", // Optional: Import GitHub keys for a specific user
    }

    res, err := client.CreateNode(hostGroup, createReq)
    if err != nil {
        panic(fmt.Errorf("failed to create node: %w", err))
    }

    fmt.Printf("Created VM: hostname=%s ip=%s created_at=%s\n", res.Hostname, res.IP, res.CreatedAt)
    fmt.Printf("Parsed IP only: %s\n", res.IPAddress())
}
```

Start Slicer:

```bash
sudo -E slicer up ./api.yaml
```

Run the program i.e. after running `go build -o client main.go`:

```bash
SLICER_URL=http://127.0.0.1:8080 SLICER_TOKEN="$(sudo cat /var/lib/slicer/auth/token)" ./client
```

You'll find the logs for the microVM at `/var/log/slicer/HOSTNAME.txt`, showing the userdata executing.

Notes:

* The argument order for `NewSlicerClient` is `(baseURL, token, userAgent, httpClient)`.
* `RAMGB`, `CPUs`, and `ImportUser` should be set; `Userdata` and `SSHKeys` are optional.
* `Userdata` runs on first boot; keep it idempotent.
* Use a persistent `http.Client` (e.g. with timeout) in production instead of `nil`.
