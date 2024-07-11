# Changelog
All notable changes to this project will be documented in this file.

**ATTN**: This project uses [semantic versioning](http://semver.org/).

## [Unreleased]

## [v0.3.0] - 2024-07-11
### Added
- Added QueryString interface method to context.
- Added QueryParams interface method to context.

### Updated
- Updated Golang version to 1.21
- Updated golangci linter to 1.55.2 version

## [v0.2.2] - 2022-11-21
### Fixed
- Fixed ALL access router. Custom permission now can be added with this route

## [v0.2.1] - 2022-11-21
### Added
- Added `access_order` param to config. Use it to specify order of channels to command access. Rights are listed in ascending order. The smallest rights come first, then the largest. You can keep it empty to use default order

## [v0.2.0] - 2022-11-21
### Changed
- Function commandHandler doesn't check default permissions anymore. It checks only selected command permissions. Added `safemode` param to config to keep old behavior

## [v0.1.3] - 2022-11-20
### Added
- Added changelog
- Added GENERAL router register

### Updated
- Updated Golang version to 1.19
- Updated golangci linter to 1.50.1 version

## [v0.1.2] - 2021-12-03
### Removed
- Removed token required validation
- Remove test channel constant

## [v0.1.1] - 2021-10-13
### Added
- Added all intents on NewSession

## [v0.1.0] - 2021-10-04
### Added
- Initial implementation

[Unreleased]: https://github.com/outdead/discordant/compare/v0.2.2...HEAD
[v0.2.2]: https://github.com/outdead/discordant/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/outdead/discordant/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/outdead/discordant/compare/v0.1.3...v0.2.0
[v0.1.3]: https://github.com/outdead/discordant/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/outdead/discordant/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/outdead/discordant/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/outdead/discordant/compare/2d21ed191dcf69520769feb1d97946600182adbc...v0.1.0
