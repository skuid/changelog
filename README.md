[![Build Status](https://travis-ci.org/skuid/aws-tag-dns.svg)](https://travis-ci.org/skuid/changelog)
[![https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](http://godoc.org/github.com/skuid/changelog/)

# changelog

changelog is a port of [clog-cli](https://github.com/clog-tool/clog-cli), with some additions. The goal is to be able to query git providers' APIs directly in addition to querying  a local git repository.

**changelog is not stable, and all interfaces, flags, and commands are subject to change until a v1.0.0 is reached**

## Installation

```
go get -u github.com/skuid/changelog
```

## Usage

```
Usage of changelog:
      --changelog string                    The Changelog file to write. Defaults to STDOUT if not set.
  -f, --from string                         The beginning commit. Defaults to beginning of the repository history
      --from-latest-tag                     If you use tags, set to true to get changes from latest tag.
      --git-dir $(pwd)/.git                 The path to the git directory. If no '--repo' is set, defaults to $(pwd)/.git. Only applies to local provider
      --include-all                         Set to true to include all commits in the changelog. Commit messages that cannot be parsed will be placed in a section titled "Unknown".
  -p, --provider string                     The provider to use. Must be one of local, github (default "local")
  -r, --repo $(git remote get-url origin)   The repository URL. Defaults to $(git remote get-url origin) if using a local provider
      --since string                        Show commits more recent than a specific date. Use RFC3339 time '2017-08-01T00:00:00Z'. Takes precedence over to/from.
      --subtitle string                     The release subtitle
  -t, --to string                           The last commit. (default "HEAD")
      --token string                        API token for remote provider. Only applies to github provider
      --until string                        Show commits older than a specific date. Defaults to current time if not set, but --since is. Takes precedence over to/from.
  -v, --version string                      The version you are creating
      --work-tree string                    The path to the directory containing the .git directory. Only applies to local provider.
```

### Examples

```bash
# Use the current working directory
changelog

# Use a different directory
changelog --work-tree /path/to/your/repo

# Query github
CHANGELOG_TOKEN="$GITHUB_ACCESS_TOKEN" changelog --repo https://github.com/skuid/changelog --provider github
```

## Configuration

All configuration options can use either environment variables with the prefix
`CHANGELOG_` or a configuration file, `.clog.toml`.

### Sections

Changelog sections may also be defined in the configuration file. All section
titles MUST be in lowercase, as [viper](https://github.com/spf13/viper) only
looks for lowercase names. The section titles will be "Title Cased" in the
final change log.

Wether using a local or remote provider, changelog will look up the `.clog.toml`
from within your repostiory and use that for config values.

```toml
[sections]
cleanup = ["cleanup", "clean"]
features = ["new"]
```

These sections are merged into the default section aliases which are:

```toml
features = ["ft", "feat"]
"bug fixes" = ["fix", "fx"]
performance = ["perf"]
"breaking changes" = ["breaks"]
unknown= ["unk"]
```

### Order

The default order of the sections in a Changelog is

* Features
* Bug Fixes
* Performance
* Breaking Changes
* Unknown

An `order` key may be set in the configuration file to specify an alternate
order. Any non-specified sections will be appended to the end alphabetically.
Any sections that don't exist will be discarded. The "Unknown" section is
always last.

## Roadmap

- [ ] Flesh out README
- [ ] Add a commit validation pre-commit hook command
- [ ] Add a web service for Github status checking (ensuring commits are properly formatted)
- [ ] Add a BitBucket Querier

## License

MIT (See [License](/LICENSE))
