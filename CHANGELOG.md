# Changelog

## [0.4.0](https://github.com/LoriKarikari/azkit/compare/azkit-v0.3.0...azkit-v0.4.0) (2026-07-05)


### Features

* add deactivate command ([8c13164](https://github.com/LoriKarikari/azkit/commit/8c131646b7e62742d92464b3397b142ea205b39c))
* add git-cliff for changelog generation ([8a46898](https://github.com/LoriKarikari/azkit/commit/8a46898b0042b4f9f822fed1db84ea0af6318513))
* add GoReleaser pipeline ([9eaff42](https://github.com/LoriKarikari/azkit/commit/9eaff425f880ad78ec4d4b4a00f5e1e79914ab49))
* add interactive progress spinners ([2120039](https://github.com/LoriKarikari/azkit/commit/21200391f9af947c685c3b2214e8e8ddd8dd11a0))
* add layered configuration with koanf ([3b6fa61](https://github.com/LoriKarikari/azkit/commit/3b6fa611cb3df43ad1c8392ed4cfd683aed6d8c5))
* add list command with fake assignments ([07ea249](https://github.com/LoriKarikari/azkit/commit/07ea249f80195decb134b5eacb2510041a21ed9f)), closes [#2](https://github.com/LoriKarikari/azkit/issues/2)
* add non-interactive activation by raw scope ID ([48ce6c6](https://github.com/LoriKarikari/azkit/commit/48ce6c61f0a6ff4e9aa67f612722f5a94d9ee7c9)), closes [#5](https://github.com/LoriKarikari/azkit/issues/5)
* add real Azure eligible assignment listing ([028fd5e](https://github.com/LoriKarikari/azkit/commit/028fd5e2767499757c6109bffd9877c7f07d42f7)), closes [#4](https://github.com/LoriKarikari/azkit/issues/4)
* add release install script ([#76](https://github.com/LoriKarikari/azkit/issues/76)) ([e6d1842](https://github.com/LoriKarikari/azkit/commit/e6d1842350aab72da1fa488f72063ccab48f7cac))
* add root version flag ([cd6babb](https://github.com/LoriKarikari/azkit/commit/cd6babbdee95fc8dc618abd162511374f00d59a2))
* add shell init support ([a00d006](https://github.com/LoriKarikari/azkit/commit/a00d006215df3675c646a8cf2b7e7e509e1fe23c)), closes [#81](https://github.com/LoriKarikari/azkit/issues/81)
* add static shell completions ([66eeb36](https://github.com/LoriKarikari/azkit/commit/66eeb3624437a6a05e8375dfd0155320914ea350))
* add status command and activation wait ([30c8079](https://github.com/LoriKarikari/azkit/commit/30c8079cc5782e4413178cd07e590867beb11681)), closes [#7](https://github.com/LoriKarikari/azkit/issues/7)
* add structured logging with global --verbose flag ([9bc42e1](https://github.com/LoriKarikari/azkit/commit/9bc42e1b3766260c39cd875b2d36037e005b10ac)), closes [#10](https://github.com/LoriKarikari/azkit/issues/10)
* add subscription and resource group selectors ([0e1ac14](https://github.com/LoriKarikari/azkit/commit/0e1ac143054a0d91b303170fbd09df7bf48ebe96)), closes [#6](https://github.com/LoriKarikari/azkit/issues/6)
* add tenant context catalog ([5fca9cd](https://github.com/LoriKarikari/azkit/commit/5fca9cd2c02a85fe87b97c77943cb49f42370af9)), closes [#82](https://github.com/LoriKarikari/azkit/issues/82)
* clean interactive cancellation ([859a9c9](https://github.com/LoriKarikari/azkit/commit/859a9c9fd2f5dd94d72fdc84064906b5b621cc07))
* finish ctx and sub output contracts ([#108](https://github.com/LoriKarikari/azkit/issues/108)) ([f35e5e2](https://github.com/LoriKarikari/azkit/commit/f35e5e2922e9531a30c1f0157aa0d431787a8258))
* implement tenant context switching ([d2d0c72](https://github.com/LoriKarikari/azkit/commit/d2d0c725513880d09aebd90750013850f7f55916)), closes [#83](https://github.com/LoriKarikari/azkit/issues/83)
* include subscription fields in PIM JSON ([5a036de](https://github.com/LoriKarikari/azkit/commit/5a036de2be7f7426b709e4b80b4c0cd54de613a3))
* interactive activation flow with Charm huh ([2428e8f](https://github.com/LoriKarikari/azkit/commit/2428e8fcebef6123b49af8115a08cd9953347403))
* interactive context and subscription pickers ([#107](https://github.com/LoriKarikari/azkit/issues/107)) ([209237c](https://github.com/LoriKarikari/azkit/commit/209237c5335b47605bcc66061e6dca9b8cffed68))
* make activation wait configurable ([ca0bc3f](https://github.com/LoriKarikari/azkit/commit/ca0bc3f7f00ed1fdac75d8385c644df7228cc064))
* rename CLI to azkit and move PIM under pim ([88b543a](https://github.com/LoriKarikari/azkit/commit/88b543ab9cded4d2721aa06eb615b9a8a67bf014)), closes [#79](https://github.com/LoriKarikari/azkit/issues/79)
* show single assignment before activation details ([dbfaaf9](https://github.com/LoriKarikari/azkit/commit/dbfaaf99d473acd9df4c580c9c2745416b57799d))
* show spinner while waiting for activation ([#51](https://github.com/LoriKarikari/azkit/issues/51)) ([c3bd0e2](https://github.com/LoriKarikari/azkit/commit/c3bd0e2459ecc4cbfbadeb5e94490941fd1fbfb9))
* skip activation when role is already active ([#52](https://github.com/LoriKarikari/azkit/issues/52)) ([23b0a8a](https://github.com/LoriKarikari/azkit/commit/23b0a8ad199e5a208d7a0408e68ca70b48eddaef))
* subscription aliases and direct switching ([#106](https://github.com/LoriKarikari/azkit/issues/106)) ([50b8a84](https://github.com/LoriKarikari/azkit/commit/50b8a844488f923b8fb761ba41cc5613d7c3a151))
* subscription discovery cache ([#105](https://github.com/LoriKarikari/azkit/issues/105)) ([6182a3a](https://github.com/LoriKarikari/azkit/commit/6182a3a2c68bea1854813db968b9ebf8731e4f05))
* switch config to azkit schema ([e5ee6c6](https://github.com/LoriKarikari/azkit/commit/e5ee6c6a79d3364bf1005130d29f43da12a66faf)), closes [#80](https://github.com/LoriKarikari/azkit/issues/80)
* tenant header in interactive forms ([a82f491](https://github.com/LoriKarikari/azkit/commit/a82f49113bff7fc15d1636091190887c2134926e))


### Bug Fixes

* align deactivation request with Azure PIM ([41e8029](https://github.com/LoriKarikari/azkit/commit/41e8029f273219ef88cd98067ed58529c34d71f0))
* classify and dedupe inherited Azure scopes ([3fc04c3](https://github.com/LoriKarikari/azkit/commit/3fc04c3d78a16a4aaad6f61b6034f20787661d97))
* clean up error classification and role matching ([22205af](https://github.com/LoriKarikari/azkit/commit/22205af4939c0f1a9160c63a477e822ffbeaa526))
* consistent activating spinner across submit and wait ([27ad157](https://github.com/LoriKarikari/azkit/commit/27ad157efa63a092d20ef5f64d0b81a1d7af2113))
* harden JSON CLI behavior ([d0ac9b1](https://github.com/LoriKarikari/azkit/commit/d0ac9b1c2738e37efca30ca909e9dc1c37022731))
* ignore active assignments without expiry ([#53](https://github.com/LoriKarikari/azkit/issues/53)) ([6f29ea1](https://github.com/LoriKarikari/azkit/commit/6f29ea196912264bed39d9056ec8c7afe96c8e80))
* point status details to extended flag ([8c1e9b0](https://github.com/LoriKarikari/azkit/commit/8c1e9b0d1c0d5a33d2bd1520b5d41b20e7fe99cb))
* remove release-as from release-please ([d3b5fbb](https://github.com/LoriKarikari/azkit/commit/d3b5fbbbd3240a11bbd90b8b70135d427b502591))
* respect release note sections ([#77](https://github.com/LoriKarikari/azkit/issues/77)) ([0321a09](https://github.com/LoriKarikari/azkit/commit/0321a097a820ee3d3b6c29859926eccd9f219413))
* return signal-style code on cancellation ([9f50d00](https://github.com/LoriKarikari/azkit/commit/9f50d00479f0451b32723942d55ce5561dad6edd))

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
