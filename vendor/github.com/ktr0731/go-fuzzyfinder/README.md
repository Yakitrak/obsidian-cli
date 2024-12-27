# go-fuzzyfinder

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ktr0731/go-fuzzyfinder)](https://pkg.go.dev/github.com/ktr0731/go-fuzzyfinder)
[![GitHub Actions](https://github.com/ktr0731/go-fuzzyfinder/workflows/main/badge.svg)](https://github.com/ktr0731/go-fuzzyfinder/actions)
[![codecov](https://codecov.io/gh/ktr0731/go-fuzzyfinder/branch/master/graph/badge.svg?token=RvpSTKDJGO)](https://codecov.io/gh/ktr0731/go-fuzzyfinder)  

`go-fuzzyfinder` is a Go library that provides fuzzy-finding with an fzf-like terminal user interface.

![](https://user-images.githubusercontent.com/12953836/52424222-e1edc900-2b3c-11e9-8158-8e193844252a.png)

## Installation
``` bash
$ go get github.com/ktr0731/go-fuzzyfinder
```

## Usage
`go-fuzzyfinder` provides two functions, `Find` and `FindMulti`.
`FindMulti` can select multiple lines. It is similar to `fzf -m`.

This is [an example](//github.com/ktr0731/go-fuzzyfinder/blob/master/example/track/main.go) of `FindMulti`.

``` go
type Track struct {
    Name      string
    AlbumName string
    Artist    string
}

var tracks = []Track{
    {"foo", "album1", "artist1"},
    {"bar", "album1", "artist1"},
    {"foo", "album2", "artist1"},
    {"baz", "album2", "artist2"},
    {"baz", "album3", "artist2"},
}

func main() {
    idx, err := fuzzyfinder.FindMulti(
        tracks,
        func(i int) string {
            return tracks[i].Name
        },
        fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
            if i == -1 {
                return ""
            }
            return fmt.Sprintf("Track: %s (%s)\nAlbum: %s",
                tracks[i].Name,
                tracks[i].Artist,
                tracks[i].AlbumName)
        }))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("selected: %v\n", idx)
}
```

The execution result prints selected item's indexes.

## Motivation
Fuzzy-finder command-line tools such that
[fzf](https://github.com/junegunn/fzf), [fzy](https://github.com/jhawthorn/fzy), or [skim](https://github.com/lotabout/skim) 
are very powerful to find out specified lines interactively.
However, there are limits to deal with fuzzy-finder's features in several cases.

First, it is hard to distinguish between two or more entities that have the same text.
In the example of ktr0731/itunes-cli, it is possible to conflict tracks such that same track names, but different artists.
To avoid such conflicts, we have to display the artist names with each track name.
It seems like the problem has been solved, but it still has the problem.
It is possible to conflict in case of same track names, same artists, but other albums, which each track belongs to.
This problem is difficult to solve because pipes and filters are row-based mechanisms, there are no ways to hold references that point list entities.

The second issue occurs in the case of incorporating a fuzzy-finder as one of the main features in a command-line tool such that [enhancd](https://github.com/b4b4r07/enhancd) or [itunes-cli](https://github.com/ktr0731/itunes-cli).
Usually, these tools require that it has been installed one fuzzy-finder as a precondition.
In addition, to deal with the fuzzy-finder, an environment variable configuration such that `export TOOL_NAME_FINDER=fzf` is required.
It is a bother and complicated.

`go-fuzzyfinder` resolves above issues.
Dealing with the first issue, `go-fuzzyfinder` provides the preview-window feature (See an example in [Usage](#usage)).
Also, by using `go-fuzzyfinder`, built tools don't require any fuzzy-finders.

## See Also
- [Fuzzy-finder as a Go library](https://medium.com/@ktr0731/fuzzy-finder-as-a-go-library-590b7458200f)
- [(Japanese) fzf ライクな fuzzy-finder を提供する Go ライブラリを書いた](https://syfm.hatenablog.com/entry/2019/02/09/120000)
