# Changelog

## [0.4.0](https://github.com/LoriKarikari/pimctl/compare/v0.3.0...v0.4.0) (2026-07-05)


### Features

* add root version flag ([cd6babb](https://github.com/LoriKarikari/pimctl/commit/cd6babbdee95fc8dc618abd162511374f00d59a2))
* add shell init support ([a00d006](https://github.com/LoriKarikari/pimctl/commit/a00d006215df3675c646a8cf2b7e7e509e1fe23c)), closes [#81](https://github.com/LoriKarikari/pimctl/issues/81)
* add tenant context catalog ([5fca9cd](https://github.com/LoriKarikari/pimctl/commit/5fca9cd2c02a85fe87b97c77943cb49f42370af9)), closes [#82](https://github.com/LoriKarikari/pimctl/issues/82)
* finish ctx and sub output contracts ([#108](https://github.com/LoriKarikari/pimctl/issues/108)) ([f35e5e2](https://github.com/LoriKarikari/pimctl/commit/f35e5e2922e9531a30c1f0157aa0d431787a8258))
* implement tenant context switching ([d2d0c72](https://github.com/LoriKarikari/pimctl/commit/d2d0c725513880d09aebd90750013850f7f55916)), closes [#83](https://github.com/LoriKarikari/pimctl/issues/83)
* include subscription fields in PIM JSON ([5a036de](https://github.com/LoriKarikari/pimctl/commit/5a036de2be7f7426b709e4b80b4c0cd54de613a3))
* interactive context and subscription pickers ([#107](https://github.com/LoriKarikari/pimctl/issues/107)) ([209237c](https://github.com/LoriKarikari/pimctl/commit/209237c5335b47605bcc66061e6dca9b8cffed68))
* rename CLI to azkit and move PIM under pim ([88b543a](https://github.com/LoriKarikari/pimctl/commit/88b543ab9cded4d2721aa06eb615b9a8a67bf014)), closes [#79](https://github.com/LoriKarikari/pimctl/issues/79)
* subscription aliases and direct switching ([#106](https://github.com/LoriKarikari/pimctl/issues/106)) ([50b8a84](https://github.com/LoriKarikari/pimctl/commit/50b8a844488f923b8fb761ba41cc5613d7c3a151))
* subscription discovery cache ([#105](https://github.com/LoriKarikari/pimctl/issues/105)) ([6182a3a](https://github.com/LoriKarikari/pimctl/commit/6182a3a2c68bea1854813db968b9ebf8731e4f05))
* switch config to azkit schema ([e5ee6c6](https://github.com/LoriKarikari/pimctl/commit/e5ee6c6a79d3364bf1005130d29f43da12a66faf)), closes [#80](https://github.com/LoriKarikari/pimctl/issues/80)


### Bug Fixes

* classify and dedupe inherited Azure scopes ([3fc04c3](https://github.com/LoriKarikari/pimctl/commit/3fc04c3d78a16a4aaad6f61b6034f20787661d97))
* clean up error classification and role matching ([22205af](https://github.com/LoriKarikari/pimctl/commit/22205af4939c0f1a9160c63a477e822ffbeaa526))
* harden JSON CLI behavior ([d0ac9b1](https://github.com/LoriKarikari/pimctl/commit/d0ac9b1c2738e37efca30ca909e9dc1c37022731))
* point status details to extended flag ([8c1e9b0](https://github.com/LoriKarikari/pimctl/commit/8c1e9b0d1c0d5a33d2bd1520b5d41b20e7fe99cb))
* return signal-style code on cancellation ([9f50d00](https://github.com/LoriKarikari/pimctl/commit/9f50d00479f0451b32723942d55ce5561dad6edd))

## [0.3.0](https://github.com/LoriKarikari/pimctl/compare/v0.2.0...v0.3.0) (2026-05-16)


### Features

* add interactive progress spinners ([2120039](https://github.com/LoriKarikari/pimctl/commit/21200391f9af947c685c3b2214e8e8ddd8dd11a0))
* add release install script ([#76](https://github.com/LoriKarikari/pimctl/issues/76)) ([e6d1842](https://github.com/LoriKarikari/pimctl/commit/e6d1842350aab72da1fa488f72063ccab48f7cac))
* clean interactive cancellation ([859a9c9](https://github.com/LoriKarikari/pimctl/commit/859a9c9fd2f5dd94d72fdc84064906b5b621cc07))
* make activation wait configurable ([ca0bc3f](https://github.com/LoriKarikari/pimctl/commit/ca0bc3f7f00ed1fdac75d8385c644df7228cc064))
* show single assignment before activation details ([dbfaaf9](https://github.com/LoriKarikari/pimctl/commit/dbfaaf99d473acd9df4c580c9c2745416b57799d))
* tenant header in interactive forms ([a82f491](https://github.com/LoriKarikari/pimctl/commit/a82f49113bff7fc15d1636091190887c2134926e))


### Bug Fixes

* align deactivation request with Azure PIM ([41e8029](https://github.com/LoriKarikari/pimctl/commit/41e8029f273219ef88cd98067ed58529c34d71f0))
* consistent activating spinner across submit and wait ([27ad157](https://github.com/LoriKarikari/pimctl/commit/27ad157efa63a092d20ef5f64d0b81a1d7af2113))
* respect release note sections ([#77](https://github.com/LoriKarikari/pimctl/issues/77)) ([0321a09](https://github.com/LoriKarikari/pimctl/commit/0321a097a820ee3d3b6c29859926eccd9f219413))

## 0.2.0 (2026-05-14)

## What's Changed
* chore: clean up deactivation internals by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/49
* feat: show spinner while waiting for activation by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/51
* feat: skip activation when role is already active by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/52
* fix: ignore active assignments without expiry by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/53
* refactor: centralize activation outcome handling by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/54
* ci: let release-please own release notes by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/55


**Full Changelog**: https://github.com/LoriKarikari/pimctl/compare/v0.1.0...v0.2.0

## 0.1.0 (2026-05-12)

## What's Changed
* feat: add list command with fake assignments by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/13
* ci: add test vet lint workflow by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/14
* feat: add real Azure eligible assignment listing by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/15
* feat: add non-interactive activation by raw scope ID by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/16
* feat: add subscription and resource group selectors by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/17
* feat: add status command for active assignments by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/18
* feat: add structured logging with global --verbose flag by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/19
* feat: add layered configuration with koanf by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/20
* feat: interactive activation flow with Charm huh by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/21
* feat: add static shell completions by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/22
* feat: add GoReleaser pipeline by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/23
* chore: pre-release polish by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/24
* fix: rename release-please config by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/25
* fix: add release-please manifest by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/26
* fix: set release-please changelog type to github by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/27
* fix: force release-please initial version to 0.1.0 by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/28
* fix: clean up release-please changelog generation by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/30
* fix: drop github changelog type from release-please by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/32
* fix: use github changelog type for PR-level entries by @LoriKarikari in https://github.com/LoriKarikari/pimctl/pull/34

## New Contributors
* @LoriKarikari made their first contribution in https://github.com/LoriKarikari/pimctl/pull/13

**Full Changelog**: https://github.com/LoriKarikari/pimctl/commits/v0.1.0
