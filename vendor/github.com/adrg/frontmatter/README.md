<h1 align="center">
  <div>
    <img src="https://raw.githubusercontent.com/adrg/adrg.github.io/master/assets/projects/frontmatter/logo.png" width="120px" alt="frontmatter logo"/>
  </div>
  frontmatter
</h1>

<h3 align="center">Go library for detecting and decoding various content front matter formats.</h3>

<p align="center">
    <a href="https://github.com/adrg/frontmatter/actions?query=workflow%3ACI">
        <img alt="Build status" src="https://github.com/adrg/frontmatter/workflows/CI/badge.svg">
    </a>
    <a href="https://codecov.io/gh/adrg/frontmatter">
        <img alt="Code coverage" src="https://codecov.io/gh/adrg/frontmatter/branch/master/graphs/badge.svg?branch=master">
    </a>
    <a href="https://pkg.go.dev/github.com/adrg/frontmatter">
        <img alt="pkg.go.dev documentation" src="https://pkg.go.dev/badge/github.com/adrg/frontmatter">
    </a>
    <a href="https://opensource.org/licenses/MIT" rel="nofollow">
        <img alt="MIT License" src="https://img.shields.io/github/license/adrg/frontmatter"/>
    </a>
    <a href="https://goreportcard.com/report/github.com/adrg/frontmatter">
        <img alt="Go report card" src="https://goreportcard.com/badge/github.com/adrg/frontmatter?style=flat" />
    </a>
    <br />
    <a href="https://github.com/adrg/frontmatter/graphs/contributors">
        <img alt="GitHub contributors" src="https://img.shields.io/github/contributors/adrg/frontmatter" />
    </a>
    <a href="https://github.com/adrg/frontmatter/issues?q=is%3Aopen+is%3Aissue">
        <img alt="GitHub open issues" src="https://img.shields.io/github/issues-raw/adrg/frontmatter">
    </a>
    <a href="https://github.com/adrg/frontmatter/issues?q=is%3Aissue+is%3Aclosed">
        <img alt="GitHub closed issues" src="https://img.shields.io/github/issues-closed-raw/adrg/frontmatter" />
    </a>
    <a href="https://www.buymeacoffee.com/adrg">
        <img alt="Buy me a coffee" src="https://img.shields.io/static/v1.svg?label=%20&message=Buy%20me%20a%20coffee&color=FF813F&logo=buy%20me%20a%20coffee&logoColor=white"/>
    </a>
    <a alt="Github stars" href="https://github.com/adrg/frontmatter/stargazers">
        <img alt="GitHub stars" src="https://img.shields.io/github/stars/adrg/frontmatter?style=social">
    </a>

## Supported formats

The following front matter formats are supported by default. If the default
formats are not suitable for your use case, you can easily bring your own.
For more information, see the [usage examples](#usage) below.

![Default front matter formats](https://raw.githubusercontent.com/adrg/adrg.github.io/master/assets/projects/frontmatter/formats.png)

## Installation

```bash
go get github.com/adrg/frontmatter
```

## Usage

**Default usage.**

```go
package main

import (
	"fmt"
	"strings"

	"github.com/adrg/frontmatter"
)

var input = `
---
name: "frontmatter"
tags: ["go", "yaml", "json", "toml"]
---
rest of the content`

func main() {
	var matter struct {
		Name string   `yaml:"name"`
		Tags []string `yaml:"tags"`
	}

	rest, err := frontmatter.Parse(strings.NewReader(input), &matter)
	if err != nil {
		// Treat error.
	}
	// NOTE: If a front matter must be present in the input data, use
	//       frontmatter.MustParse instead.

	fmt.Printf("%+v\n", matter)
	fmt.Println(string(rest))

	// Output:
	// {Name:frontmatter Tags:[go yaml json toml]}
	// rest of the content
}
```

**Bring your own formats.**

```go
package main

import (
	"fmt"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

var input = `
...
name: "frontmatter"
tags: ["go", "yaml", "json", "toml"]
...
rest of the content`

func main() {
	var matter struct {
		Name string   `yaml:"name"`
		Tags []string `yaml:"tags"`
	}

	formats := []*frontmatter.Format{
		frontmatter.NewFormat("...", "...", yaml.Unmarshal),
	}

	rest, err := frontmatter.Parse(strings.NewReader(input), &matter, formats...)
	if err != nil {
		// Treat error.
	}
	// NOTE: If a front matter must be present in the input data, use
	//       frontmatter.MustParse instead.

	fmt.Printf("%+v\n", matter)
	fmt.Println(string(rest))

	// Output:
	// {Name:frontmatter Tags:[go yaml json toml]}
	// rest of the content
}
```

Full documentation can be found at: https://pkg.go.dev/github.com/adrg/frontmatter.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/adrg/frontmatter.svg)](https://starchart.cc/adrg/frontmatter)

## Contributing

Contributions in the form of pull requests, issues or just general feedback,
are always welcome.
See [CONTRIBUTING.MD](CONTRIBUTING.md).

## Buy me a coffee

If you found this project useful and want to support it, consider buying me a coffee.  
<a href="https://www.buymeacoffee.com/adrg">
    <img src="https://cdn.buymeacoffee.com/buttons/v2/arial-orange.png" alt="Buy Me A Coffee" height="42px">
</a>

## License

Copyright (c) 2020 Adrian-George Bostan.

This project is licensed under the [MIT license](https://opensource.org/licenses/MIT).
See [LICENSE](LICENSE) for more details.
