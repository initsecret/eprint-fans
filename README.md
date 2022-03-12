# `eprint-fans`

[![Go](https://github.com/sgmenda/eprint-fans/actions/workflows/go.yml/badge.svg)](https://github.com/sgmenda/eprint-fans/actions/workflows/go.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](/LICENSE)

_Generate custom ePrint feeds using keywords._

Live at [eprint.fans](https://eprint.fans).

**Why does it include a cursed RSS parser?** ePrint's RSS feed includes the unsanitized content of paper abstracts, which frequently leads to the feed being malformed XML,[^why-cursed] so it is hard parse with run-of-the-mill RSS parsers---even ones that try to work with malformed feeds like [gofeed](https://github.com/mmcdole/gofeed).

[^why-cursed]: Yup, I reached out to the eprint admins, but they didn't get back to me, so I went the cursed parser route. I swear, I don't like writing cursed parsers.

**Running your own instance.** Since this is a simple Go app, it should be easy to run your own instance. I am using fly which makes it [super easy](https://fly.io/docs/getting-started/golang/). Other one-click container services like [aws copilot](https://aws.amazon.com/containers/copilot/) should also be super easy (you might need to write a standard Go Dockerfile.)

---

#### License

`eprint-fans` is licensed under [Apache 2.0](/LICENSE).
