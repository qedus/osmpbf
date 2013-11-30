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
	f, err := os.Open("planet-latest.osm.pbf")
	if err != nil {
		return err
	}
	defer f.Close()

	d := pbf.NewDecoder(f)
	n, w, r := 0, 0 , 0
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err.Error())
		} else {
			switch v := v.(type) {
			case *pbf.Node:
				// Process Node v.
				n++
			case *pbf.Way:
				// Process Way v.
				w++
			case *pbf.Relation:
				// Process Relation v.
				r++
			default:
				return fmt.Errorf("unknwon type %T\n", v)
			}
		}
	}
	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", n, w, r)
```

## Performance

My old 2.53 GHz Intel Core 2 Duo MacBook Pro with a SATA hard drive can run the above program over the whole planet as of late 2013 in just over 1 hour. The code is probably more disk IO bound than CPU bound though.

## To Do

The parseNodes code has not been tested as I can only find PBF files with DenseNode format.

The code does not decode DenseInfo or Info data structures as I currently have no need for them.

An Encoder still needs to be created to reverse the process.
