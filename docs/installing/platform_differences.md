# Understanding Platform Differences

Some system collectors are specific to one platform. On macOS, radar skips the
Linux-only collectors for systemd, SELinux, and the `/proc` and `/sys`
filesystems. radar uses macOS collectors such as `sysctl`, `system_profiler`,
`diskutil`, and `launchctl` instead.
