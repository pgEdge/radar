# Downloading a Pre-built Binary

Each [GitHub release](https://github.com/pgEdge/radar/releases) includes static
radar binaries for Linux and macOS on amd64 and arm64. The release also
includes a `checksums.txt` file with the SHA-256 checksum of each binary.

Use the following steps to install a pre-built binary on Linux:

1. Download the binary for your architecture and the checksum file from the
   latest release; the following commands download the amd64 binary:

    ```bash
    curl -fLO https://github.com/pgEdge/radar/releases/latest/download/radar-linux-amd64
    curl -fLO https://github.com/pgEdge/radar/releases/latest/download/checksums.txt
    ```

2. Verify the binary against the published checksum; the command prints `OK`
   after the name of a matching binary:

    ```bash
    sha256sum --check --ignore-missing checksums.txt
    ```

3. Make the binary executable and move the binary into a directory on your
   `PATH`:

    ```bash
    chmod +x radar-linux-amd64
    sudo mv radar-linux-amd64 /usr/local/bin/radar
    ```

On macOS, download `radar-darwin-arm64` or `radar-darwin-amd64` instead. Run
`shasum -a 256` on the binary, and compare the output with the matching line in
`checksums.txt`.
