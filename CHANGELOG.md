# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed
- Critical thresholds are now evaluated before warning thresholds. Each branch
  returns as soon as it matches, so with the expected `warn < crit` ordering the
  warning branch always matched first and the check could never report critical.

### Added
- Real unit tests for `checkArgs`, `Round` and `executeCheck`, covering the
  threshold ordering, the `--cmdline` matching mode and the opt-in runtime
  thresholds.
- A README documenting what the check actually does, its arguments and its
  annotation keyspace, replacing the leftover check-plugin-template boilerplate.

### Changed
- Updated the Go toolchain to 1.26 and refreshed all module dependencies,
  including `github.com/sensu/core/v2` to 2.21.5.
- Bumped the GitHub Actions to their current majors: `actions/checkout` v7,
  `actions/setup-go` v7, `golangci/golangci-lint-action` v9 and
  `goreleaser/goreleaser-action` v7.
- `release.yml` now uses `fetch-depth: 0` on checkout instead of a separate
  `git fetch --unshallow` step.
- Migrated `.goreleaser.yml` to the GoReleaser v2 schema: added `version: 2`,
  replaced the deprecated `archives.format` with `archives.formats`, and removed
  the `goos`/`goarch`/`goarm` lists that are ignored when `builds.targets` is set.

### Removed
- The `tag.yml` workflow, which was no longer used to cut releases.

### Security
- Bumped `golang.org/x/net` from 0.34.0 to 0.38.0 and
  `github.com/golang-jwt/jwt/v4` from 4.5.1 to 4.5.2 to pick up advisory fixes.

## [1.2.3] - 2025-01-10

### Changed
- Updated the Go version and module dependencies.
- Switched the release workflow to `goreleaser-action` v6.
- Widened the dependabot configuration.

### Fixed
- Corrected the argument usage strings and the output formatting.

## [1.2.2] - 2024-04-08

### Added
- The release workflow can be started manually with `workflow_dispatch`.
- A tag workflow for cutting releases from the GitHub UI.
- Dependabot configuration for Go modules and GitHub Actions.

### Changed
- Updated the Go version and module dependencies.

## [1.2.1] - 2023-11-06

### Fixed
- The runtime critical threshold returned a warning exit status instead of a
  critical one.

## [1.2] - 2023-11-06

### Added
- `--cmdline`, to match a process on its full command line rather than on its
  name.

## [1.1] - 2023-11-06

### Added
- `--time-warn` and `--time-crit`, to alert on how long a process has been
  running. Both are opt-in: a value of `0` disables the check.

## [1.0.2] - 2023-10-12

### Fixed
- The CPU thresholds were registered under the wrong argument names.

## [1.0.1] - 2023-10-12

### Added
- Initial release: check the memory and CPU usage of a named process against
  warning and critical thresholds.

[1.2.3]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.2.3
[1.2.2]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.2.2
[1.2.1]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.2.1
[1.2]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.2
[1.1]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.1
[1.0.2]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.0.2
[1.0.1]: https://github.com/elfranne/sensu-process-ressources/releases/tag/1.0.1
