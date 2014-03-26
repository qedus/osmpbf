package osmpbf

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	// originally downloaded from http://download.geofabrik.de/europe/great-britain/england/greater-london.html
	London    = "greater-london-140324.osm.pbf"
	LondonURL = "https://googledrive.com/host/0B8pisLiGtmqDR3dOR3hrWUpRTVE"
)

func init() {
	_, err := os.Stat(London)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("\nDownload %s from %s.\nFor example: 'wget -O %s %s'", London, LondonURL, London, LondonURL))
	}
}

func TestDecoder(t *testing.T) {
	f, err := os.Open(London)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var n *Node
	en := &Node{
		ID:  18088578,
		Lat: 51.5442632,
		Lon: -0.2010027,
		Tags: map[string]string{
			"alt_name":   "The King's Head",
			"amenity":    "pub",
			"created_by": "JOSM",
			"name":       "The Luminaire",
			"note":       "Live music venue too",
		},
	}

	var w *Way
	ew := &Way{
		ID:      4257116,
		NodeIDs: []int64{21544864, 333731851, 333731852, 333731850, 333731855, 333731858, 333731854, 108047, 769984352, 21544864},
		Tags: map[string]string{
			"area":    "yes",
			"highway": "pedestrian",
			"name":    "Fitzroy Square",
		},
	}

	var r *Relation
	er := &Relation{
		ID: 7677,
		Members: []Member{
			Member{ID: 4875932, Type: WayType, Role: "outer"},
			Member{ID: 4894305, Type: WayType, Role: "inner"},
		},
		Tags: map[string]string{
			"created_by": "Potlatch 0.9c",
			"type":       "multipolygon",
		},
	}

	d := NewDecoder(f)
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		} else {
			switch v := v.(type) {
			case *Node:
				if v.ID == en.ID {
					n = v
				}
			case *Way:
				if v.ID == ew.ID {
					w = v
				}
			case *Relation:
				if v.ID == er.ID {
					r = v
				}
			default:
				t.Fatalf("unknown type %T", v)
			}
		}
	}

	if !reflect.DeepEqual(en, n) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", en, n)
	}
	if !reflect.DeepEqual(ew, w) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", ew, w)
	}
	if !reflect.DeepEqual(er, r) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", er, r)
	}
}

func BenchmarkDecoder(b *testing.B) {
	f, err := os.Open(London)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)
		d := NewDecoder(f)
		n, w, r, count, start := 0, 0, 0, 0, time.Now()
		for {
			if v, err := d.Decode(); err == io.EOF {
				break
			} else if err != nil {
				b.Fatal(err)
			} else {
				switch v := v.(type) {
				case *Node:
					n++
				case *Way:
					w++
				case *Relation:
					r++
				default:
					b.Fatalf("unknown type %T", v)
				}
			}
			count++
		}
		b.Logf("Done in %.3f seconds. Total: %d, Nodes: %d, Ways: %d, Relations: %d\n",
			time.Now().Sub(start).Seconds(), count, n, w, r)
	}
}
