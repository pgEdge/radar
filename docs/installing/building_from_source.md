# Building from Source

Building radar from source requires Git and Go 1.26.10 or later. The
[Go installation guide](https://go.dev/doc/install) describes how to install a
current Go release on any platform.

Use the following steps to build radar from source:

1. Confirm that `go version` reports Go 1.26.10 or later:

    ```bash
    go version
    ```

2. Clone the pgEdge radar repository and change into its directory:

    ```bash
    git clone https://github.com/pgEdge/radar.git
    cd radar
    ```

3. Build a static binary for your platform:

    ```bash
    CGO_ENABLED=0 go build -ldflags="-s -w" -o radar .
    ```

4. Optionally, move the binary into a directory on your `PATH`:

    ```bash
    sudo mv radar /usr/local/bin/
    ```

The
[CONTRIBUTING.md](https://github.com/pgEdge/radar/blob/main/CONTRIBUTING.md)
file describes cross-compilation and the development setup.
