## Install CodexDB CLI (`codexdb`)

### From GitHub Releases (recommended)

- Download the `codexdb_<version>_<os>_<arch>.tar.gz` (or `.zip` on Windows) asset from the latest release.
- Verify against `checksums.txt`.
- Put `codexdb` on your `PATH`.

### Homebrew (macOS/Linux)

Once the Homebrew tap is published:

```bash
brew tap evertonmj/tap
brew install codexdb
codexdb --version
```

The shorter alias is also available:

```bash
cdx --version
```

### apt (Debian/Ubuntu)

This repo’s releases include a `.deb` package built by CI.

#### Install from a downloaded `.deb`

```bash
sudo apt install ./codexdb_<version>_<arch>.deb
codexdb --version
```

#### Install from an apt repository (optional)

To support `apt install codexdb` directly, publish an apt repository (e.g. via GitHub Pages) and sign it with GPG.
If you want, I can scaffold the `gh-pages` publishing workflow and repository layout next.

