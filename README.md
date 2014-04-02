osmpbf
===

Package osmpbf is used to decode OpenStreetMap pbf files.

## Installation

```bash
$ go get github.com/qedus/osmpbf
```

## Usage

Usage is similar to `json.Decode`.

```Go
	f, err := os.Open("greater-london-140324.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := osmpbf.NewDecoder(f)
	err = d.Start(runtime.GOMAXPROCS(-1)) // use several goroutines for faster decoding
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

http://godoc.org/github.com/qedus/osmpbf

## Performance

My old 2.53 GHz Intel Core 2 Duo MacBook Pro with a SATA hard drive can run the above program over the whole planet as of late 2013 in just over 1 hour. The code is probably more disk IO bound than CPU bound.

## To Do

The parseNodes code has not been tested as I can only find PBF files with DenseNode format.

The code does not decode DenseInfo or Info data structures as I currently have no need for them.

An Encoder still needs to be created to reverse the process.
