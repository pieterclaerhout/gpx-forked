# go-gpx

[![Travis CI status](https://api.travis-ci.org/thcyron/go-gpx.svg)](https://travis-ci.org/thcyron/go-gpx)

`go-gpx` is a Go library for parsing GPX 1.1 documents.

# Installation

    go get github.com/thcyron/go-gpx

# Usage

```go
f, err := os.Open("test.gpx")
if err != nil {
        panic(err)
}

doc, err := gpx.NewDecoder(f).Decode()
if err != nil {
        panic(err)
}

fmt.Printf("document has %d track(s)\n", len(doc.Tracks))
```

# Documentation

Documentation is available at [Godoc](http://godoc.org/github.com/thcyron/go-gpx).

# License

`go-gpx` is licensed under the MIT license.
