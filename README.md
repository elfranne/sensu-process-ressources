[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/elfranne/sensu-process-ressources)
![Go Test](https://github.com/elfranne/sensu-process-ressources/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/elfranne/sensu-process-ressources/workflows/goreleaser/badge.svg)

# sensu-process-ressources

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
  - [Examples](#examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
  - [Annotations](#annotations)
- [Configuration reference](#configuration-reference)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

sensu-process-ressources is a [Sensu Check][6] that watches the resource usage of
a named process and alerts when it goes over a threshold. It can alert on three
things, in any combination:

- **memory** — resident memory as a percentage of total system memory
- **CPU** — CPU usage as a percentage, averaged over the lifetime of the process
- **runtime** — how long the process has been running, in seconds

Every running process is examined, and the check reports on the first one whose
name (or full command line, with `--cmdline`) matches `--process`. If several
processes share a name, the first match wins rather than the busiest one.

The check exits `0` (OK) when nothing matched a threshold — including when no
process matched the name at all. It does not alert on a missing process; use a
dedicated process-presence check for that.

## Usage examples

### Help output

```
Check if process is using too much ressources (CPU/memory)

Usage:
  check-process-ressources [flags]
  check-process-ressources [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
      --cmdline               Use full command line of the process
      --cpu-crit float        Critical if process is using more than cpu-crit (in percent) (default 75)
      --cpu-warn float        Warn if process is using more than cpu-warn (in percent) (default 50)
  -h, --help                  help for check-process-ressources
      --memory-crit float32   Critical if process is using more than memory-crit (in percent) (default 70)
      --memory-warn float32   Warn if process is using more than memory-warn (in percent) (default 50)
      --process string        Process to monitor
      --time-crit int         Critical if process has been running for longer than time-crit (in seconds)
      --time-warn int         Warn if process has been running for longer than time-warn (in seconds)

Use "check-process-ressources [command] --help" for more information about a command.
```

### Examples

Alert when `nginx` uses more than 60% of memory (critical) or 40% (warning),
leaving the CPU thresholds at their defaults:

```
sensu-process-ressources --process nginx --memory-warn 40 --memory-crit 60
```

Alert on a long-running backup job. The memory and CPU thresholds are raised out
of the way so that only the runtime triggers:

```
sensu-process-ressources --process backup.sh \
  --memory-warn 95 --memory-crit 99 \
  --cpu-warn 95 --cpu-crit 99 \
  --time-warn 3600 --time-crit 7200
```

Match on the full command line instead of the process name, which is how you tell
two processes of the same binary apart:

```
sensu-process-ressources --cmdline --process "/usr/bin/python3 /opt/app/worker.py --queue email"
```

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add elfranne/sensu-process-ressources
```

If you're using an earlier version of sensuctl, you can find the asset on the
[Bonsai Asset Index](https://bonsai.sensu.io/assets/elfranne/sensu-process-ressources).

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: sensu-process-ressources
  namespace: default
spec:
  command: sensu-process-ressources --process nginx --memory-warn 40 --memory-crit 60
  subscriptions:
  - system
  runtime_assets:
  - elfranne/sensu-process-ressources
```

### Annotations

Every argument can also be set as an entity or check annotation under the
`sensu.io/plugins/check-process-ressources/config` keyspace, which lets you keep
one check definition and vary the thresholds per entity:

```yml
---
type: Entity
api_version: core/v2
metadata:
  name: web-01
  namespace: default
  annotations:
    sensu.io/plugins/check-process-ressources/config/process: nginx
    sensu.io/plugins/check-process-ressources/config/memory-crit: "60"
```

## Configuration reference

| Argument        | Type    | Default | Description |
| --------------- | ------- | ------- | ----------- |
| `--process`     | string  | —       | Process to match. Required. |
| `--cmdline`     | bool    | `false` | Match against the full command line instead of the process name. |
| `--memory-warn` | float32 | `50`    | Warn above this percentage of system memory. |
| `--memory-crit` | float32 | `70`    | Go critical above this percentage of system memory. |
| `--cpu-warn`    | float64 | `50`    | Warn above this CPU percentage. |
| `--cpu-crit`    | float64 | `75`    | Go critical above this CPU percentage. |
| `--time-warn`   | int64   | `0`     | Warn when the process has run this many seconds. `0` disables the check. |
| `--time-crit`   | int64   | `0`     | Go critical when the process has run this many seconds. `0` disables the check. |

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-process-ressources repository:

```
go build
```

To run the tests:

```
go test ./...
```

## Additional notes

- **Thresholds are checked in a fixed order**: memory critical, memory warning,
  CPU critical, CPU warning, runtime critical, runtime warning. The first one
  that matches decides the exit status, so a process over both its memory and
  CPU limits is reported as a memory problem.
- **A threshold of exactly `100` is rejected.** The check exits with a warning
  and an error rather than running, because a percentage that can never be
  exceeded is almost always a mistake.
- **CPU percentages can exceed 100** on a multi-core machine, since they are
  measured per core. A process saturating two cores reports roughly 200%.
- **CPU usage is a lifetime average**, not an instantaneous sample. A process
  that is busy right now but has been idle for hours will report a low value.
- **The runtime thresholds are opt-in.** They are only applied when set above
  `0`, so leaving them out means the check never alerts on process age.

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[6]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
