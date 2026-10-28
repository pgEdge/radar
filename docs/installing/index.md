# Installing radar

radar runs on Linux and macOS as a single static binary. You can install the
pgEdge package, download a pre-built binary, or build radar from source.

## Reviewing the Requirements

The requirements for radar include:

- a Linux or macOS host; radar uses standard utilities such as `df` and `ps`
  when they are present.
- PostgreSQL 12 or later; some collectors require a later version, as the
  [data reference](../collector_reference/index.md) notes.
- Git and Go 1.25.12 or later for a build from source; the `go` directive in
  [go.mod](https://github.com/pgEdge/radar/blob/main/go.mod) sets this minimum.

## Understanding Platform Differences

Some system collectors are specific to one platform. On macOS, radar skips the
Linux-only collectors for systemd, SELinux, and the `/proc` and `/sys`
filesystems. radar uses macOS collectors such as `sysctl`, `system_profiler`,
`diskutil`, and `launchctl` instead.
