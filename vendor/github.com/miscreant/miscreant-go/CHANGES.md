## [0.3.0] (2017-12-25)

[0.3.0]: https://github.com/miscreant/miscreant/compare/v0.2.0...v0.3.0

* STREAM support (all languages)
* AEAD APIs: TypeScript, Rust
* Rust internals based on RustCrypto project providing ~10% faster performance

### Notable Pull Requests

* [#100](https://github.com/miscreant/miscreant/pull/100)
  rust: Use AES-CTR implementation from the `aesni` crate
* [#102](https://github.com/miscreant/miscreant/pull/102)
  rust: Use cmac and pmac crates from RustCrypto
* [#103](https://github.com/miscreant/miscreant/pull/103)
  rust: DRY out and abstract SIV+CTR implementations across key sizes
* [#104](https://github.com/miscreant/miscreant/pull/104)
  rust: Deny unsafe_code
* [#105](https://github.com/miscreant/miscreant/pull/105)
  rust: AEAD API
* [#112](https://github.com/miscreant/miscreant/pull/112)
  rust: STREAM implementation
* [#117](https://github.com/miscreant/miscreant/pull/117)
  rust: "std" feature and allocating APIs
* [#120](https://github.com/miscreant/miscreant/pull/120)
  rust: Dual license under MIT/Apache 2.0
* [#122](https://github.com/miscreant/miscreant/pull/122)
  ruby: STREAM implementation
* [#124](https://github.com/miscreant/miscreant/pull/124)
  python: STREAM implementation
* [#126](https://github.com/miscreant/miscreant/pull/126)
  go: Switch to using math.TrailingZeros in PMAC (requires Go 1.9+)
* [#127](https://github.com/miscreant/miscreant/pull/127)
  js: AEAD API
* [#131](https://github.com/miscreant/miscreant/pull/131)
  js: STREAM implementation
* [#132](https://github.com/miscreant/miscreant/pull/132)
  go: STREAM implementation

## [0.2.0] (2017-10-01)

[0.2.0]: https://github.com/miscreant/miscreant/compare/v0.1.0...v0.2.0

* AES-PMAC-SIV support (all languages)
* AEAD APIs with test vectors: Go, Ruby, Python
* Various breaking API changes from 0.1.0, but hopefully no one was using a v0.1
  crypto library anyway.

# 0.1.0 (2017-07-31)

* Initial release
