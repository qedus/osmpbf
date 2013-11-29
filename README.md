pbf
===

Package pbf is used to decode OpenStreetMap pbf files.

## Installation

```bash
$ go get github.com/qedus/pdf
```

## Usage

Usage is similar to `json.Decode`.

```Go
	// Open an io.Reader stream
	f, err := os.Open("planet-latest.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dec := pbf.NewDecoder(f)
	for {
		if v, err := dec.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case pbf.Node:
				// Handle an OSM Node
			case pbf.Way:
				// Handle an OSM Way
			case pbf.Relation:
				// Handle an OSM Relation
			}	
		}
	}
```
