# `eprint-fans`

[![Go](https://github.com/sgmenda/eprint-fans/actions/workflows/go.yml/badge.svg)](https://github.com/sgmenda/eprint-fans/actions/workflows/go.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](/LICENSE)

_Generate custom ePrint feeds using keywords._

Live at [eprint.fans](https://eprint.fans).

**Parsing the RSS feed** Uses [gofeed](https://github.com/mmcdole/gofeed) to parse [ePrint's Atom Feed](https://eprint.iacr.org/rss/atom.xml).

**Running your own instance.** Since this is a simple Go app, it should be easy to run your own instance. I am using fly which makes it [super easy](https://fly.io/docs/getting-started/golang/). Other one-click container services like [aws copilot](https://aws.amazon.com/containers/copilot/) should also be super easy (you might need to write a standard Go Dockerfile.)

---

#### License

`eprint-fans` is licensed under [Apache 2.0](/LICENSE).
