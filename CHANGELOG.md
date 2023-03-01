# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [5.0.0] - 2023-03-01
### Changed
- Response.GetTransport() now returns a Transport pointer

### Fixed
- Action properties are now properly assigned to transport
- Parameter schema changed types for "items", "max" and "min"
- Response reply payload does not leak the call data into the payload anymore
- Request initializes the attributes on demand when there are no attributes
- Change service processor to return action errors as transport errors instead of error replies
- Changed local file token validation to fail when there is a token

## [4.0.0] - 2022-03-01
### Added
- Support for the component address CLI "--address" argument

### Changed
- Change incoming socket HWM to avoid limiting the number of incoming requests
- CLI --socket option is now called --ipc
- Runtime calls now get the component address from the CLI arguments instead of the mappings
- Log level ERROR is used by default until the level is read from the CLI arguments

### Fixed
- Server socket initialization issue during error checking

## [3.0.0] - 2021-03-01
### Changed
- Complete SDK rewrite

## [2.0.0] - 2020-03-01
### Changed
- Updates version to 2.0.0
