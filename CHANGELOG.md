# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


- **[Unreleased]** - Features merged but not yet released
- **Version numbers** - Released versions with dates
- **Added** - New features
- **Changed** - Changes to existing functionality
- **Deprecated** - Features that will be removed
- **Removed** - Features that were removed
- **Fixed** - Bug fixes
- **Security** - Security-related changes

## v1.1.0 (2026-01-19)

### Feat

- add sort cycling with 's' key

### Fix

- update golangci-lint config for v2 compatibility
- correct failing test expectations in actions_test.go
- pin golangci-lint to v2.8 for config compatibility
- correct YAML structure in CI workflow
- update CI badge to explicitly reference main branch
- prevent panic after creating new project
- enable text input in new project view

## v1.0.0 (2026-01-19)

### Feat

- add Makefile targets for automated releases
- add automated release script with commitizen
- add lint-fix target to Makefile

### Fix

- handle file.Close errors in script detection
- simplify golangci-lint config for v2.8.0 compatibility
- remove ANSI escape codes from shell integration display in install script

### Refactor

- remove unused dimStyle variable
- remove unused actionDescStyle variable
- remove unused dockerfilePatterns variable
