# osmpbf

[![Build Status](https://github.com/qedus/osmpbf/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/qedus/osmpbf/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/qedus/osmpbf/badge.svg?branch=master)](https://coveralls.io/github/qedus/osmpbf?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/qedus/osmpbf)](https://goreportcard.com/report/github.com/qedus/osmpbf)
[![Go Reference](https://pkg.go.dev/badge/github.com/qedus/osmpbf.svg)](https://pkg.go.dev/github.com/qedus/osmpbf)

Package osmpbf is used to decode OpenStreetMap pbf files.

## Installation

```bash
$ go get github.com/qedus/osmpbf
```

## Usage

Usage is similar to `json.Decoder`.

```Go
	f, err := os.Open("greater-london-140324.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := osmpbf.NewDecoder(f)

	// use more memory from the start, it is faster
	d.SetBufferSize(osmpbf.MaxBlobSize)

	// start decoding with several goroutines, it is faster
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		log.Fatal(err)
	}

	var nc, wc, rc uint64
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Process Node v.
				nc++
			case *osmpbf.Way:
				// Process Way v.
				wc++
			case *osmpbf.Relation:
				// Process Relation v.
				rc++
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}

	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)
```

## Documentation

https://pkg.go.dev/github.com/qedus/osmpbf

## To Do

The parseNodes code has not been tested as I can only find PBF files with DenseNode format.

An Encoder still needs to be created to reverse the process.
