# Linux DVB Exporter

[![CI](https://github.com/coord-e/linux_dvb_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/coord-e/linux_dvb_exporter/actions/workflows/ci.yml)
[![Release](https://github.com/coord-e/linux_dvb_exporter/actions/workflows/release.yml/badge.svg)](https://github.com/coord-e/linux_dvb_exporter/actions/workflows/release.yml)

Prometheus exporter for DVB device metrics. Currently [frontend statistics](https://www.kernel.org/doc/html/v5.10/userspace-api/media/dvb/frontend-stat-properties.html#frontend-stat-properties) and [status](https://www.kernel.org/doc/html/v5.10/userspace-api/media/dvb/fe-read-status.html) are exported.
Pre-built binaries are available at [the releases](https://github.com/coord-e/linux_dvb_exporter/releases).

## Usage

```shell
$ ./linux_dvb_exporter -h
usage: linux_dvb_exporter [<flags>]

Flags:
  -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      --web.config.file=""  [EXPERIMENTAL] Path to configuration file that can enable TLS or authentication.
      --web.listen-address=":9111"
                            The address to listen on for HTTP requests.
      --web.telemetry-path="/metrics"
                            Path under which to expose metrics.
      --log.level=info      Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt   Output format of log messages. One of: [logfmt, json]
      --version             Show application version.
```

## Build

```shell
$ make build
```
