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

### Outputs of a sample code

```sh
$ go run osmpbf.go -ncpu 4 greater-london-140324.osm.pbf 
33728 / 33728 [====================================================] 100.00 % 1s
Nodes: 2,729,006, Ways: 459,055, Relations: 12,833
$ go run osmpbf.go -ncpu 1 greater-london-140324.osm.pbf 
33728 / 33728 [====================================================] 100.00 % 4s
Nodes: 2,729,006, Ways: 459,055, Relations: 12,833
```

## Documentation

http://godoc.org/github.com/qedus/osmpbf

## To Do

The parseNodes code has not been tested as I can only find PBF files with DenseNode format.

The code does not decode DenseInfo or Info data structures as I currently have no need for them.

An Encoder still needs to be created to reverse the process.
