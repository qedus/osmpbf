package osmpbf_test

import (
	"fmt"
	"github.com/qedus/osmpbf"
	"io"
	"log"
	"os"
	"runtime"
)

// Don't forget to sync with README.md
func Example() {
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

	n, w, r := 0, 0, 0
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				// Process Node v.
				n++
			case *osmpbf.Way:
				// Process Way v.
				w++
			case *osmpbf.Relation:
				// Process Relation v.
				r++
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}

	fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", n, w, r)
	// Output:
	// Nodes: 2729006, Ways: 459055, Relations: 12833
}
