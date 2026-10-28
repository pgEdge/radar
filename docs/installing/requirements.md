# Reviewing the Requirements

The requirements for radar include:

- a Linux or macOS host; radar uses standard utilities such as `df` and `ps`
  when they are present.
- PostgreSQL 12 or later; some collectors require a later version, as the
  [data reference](../collector_reference/index.md) notes.
- Git and Go 1.25.12 or later for a build from source; the `go` directive in
  [go.mod](https://github.com/pgEdge/radar/blob/main/go.mod) sets this minimum.
