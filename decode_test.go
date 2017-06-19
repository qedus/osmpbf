package osmpbf

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	// Originally downloaded from http://download.geofabrik.de/europe/great-britain/england/greater-london.html
	// Stored at https://gist.github.com/AlekSi/d4369aa13cf1fc5ddfac3e91b67b2f7b
	London    = "greater-london-140324.osm.pbf"
	LondonURL = "https://gist.githubusercontent.com/AlekSi/d4369aa13cf1fc5ddfac3e91b67b2f7b/raw/f87959d6c9466547d9759971e071a15049b67ae2/greater-london-140324.osm.pbf"
)

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

var (
	IDsExpectedOrder = []string{
		// Start of dense nodes.
		"node/44", "node/47", "node/52", "node/58", "node/60",
		"node/79", // Just because way/79 is already there
		"node/2740703694", "node/2740703695", "node/2740703697",
		"node/2740703699", "node/2740703701",
		// End of dense nodes.

		// Start of ways.
		"way/73", "way/74", "way/75", "way/79", "way/482",
		"way/268745428", "way/268745431", "way/268745434", "way/268745436",
		"way/268745439",
		// End of ways.

		// Start of relations.
		"relation/69", "relation/94", "relation/152", "relation/245",
		"relation/332", "relation/3593436", "relation/3595575",
		"relation/3595798", "relation/3599126", "relation/3599127",
		// End of relations
	}

	IDs map[string]bool

	enc uint64 = 2729006
	ewc uint64 = 459055
	erc uint64 = 12833

	eh = &Header{
		BoundingBox: &BoundingBox{
			Right:  0.335437,
			Left:   -0.511482,
			Bottom: 51.28554,
			Top:    51.69344,
		},
		OsmosisReplicationTimestamp: time.Date(2014, 3, 24, 22, 55, 2, 0, time.FixedZone("test", 3600)),
		RequiredFeatures: []string{
			"OsmSchema-V0.6",
			"DenseNodes",
		},
		WritingProgram: `Osmium (http:\/\/wiki.openstreetmap.org\/wiki\/Osmium)`,
	}

	en = &Node{
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
		Info: Info{
			Version:   2,
			Timestamp: parseTime("2009-05-20T10:28:54Z"),
			Changeset: 1260468,
			Uid:       508,
			User:      "Welshie",
			Visible:   true,
		},
	}

	ew = &Way{
		ID: 4257116,
		NodeIDs: []int64{
			21544864, 333731851, 333731852, 333731850, 333731855,
			333731858, 333731854, 108047, 769984352, 21544864},
		Tags: map[string]string{
			"area":    "yes",
			"highway": "pedestrian",
			"name":    "Fitzroy Square",
		},
		Info: Info{
			Version:   7,
			Timestamp: parseTime("2013-08-07T12:08:39Z"),
			Changeset: 17253164,
			Uid:       1016290,
			User:      "Amaroussi",
			Visible:   true,
		},
	}

	er = &Relation{
		ID: 7677,
		Members: []Member{
			{ID: 4875932, Type: WayType, Role: "outer"},
			{ID: 4894305, Type: WayType, Role: "inner"},
		},
		Tags: map[string]string{
			"created_by": "Potlatch 0.9c",
			"type":       "multipolygon",
		},
		Info: Info{
			Version:   4,
			Timestamp: parseTime("2008-07-19T15:04:03Z"),
			Changeset: 540201,
			Uid:       3876,
			User:      "Edgemaster",
			Visible:   true,
		},
	}
)

func init() {
	IDs = make(map[string]bool)
	for _, id := range IDsExpectedOrder {
		IDs[id] = false
	}
}

func downloadTestOSMFile(t *testing.T) {
	if _, err := os.Stat(London); os.IsNotExist(err) {
		resp, err := http.Get(LondonURL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		out, err := os.Create(London)
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
}

func checkHeader(a *Header) bool {
	if a == nil || a.BoundingBox == nil || a.RequiredFeatures == nil {
		return false
	}

	// check bbox
	if a.BoundingBox.Right != eh.BoundingBox.Right || a.BoundingBox.Left != eh.BoundingBox.Left || a.BoundingBox.Top != eh.BoundingBox.Top || a.BoundingBox.Bottom != eh.BoundingBox.Bottom {
		return false
	}

	// check timestamp
	if !a.OsmosisReplicationTimestamp.Equal(eh.OsmosisReplicationTimestamp) {
		return false
	}

	// check writing program
	if a.WritingProgram != eh.WritingProgram {
		return false
	}

	// check features
	if len(a.RequiredFeatures) != len(eh.RequiredFeatures) || a.RequiredFeatures[0] != eh.RequiredFeatures[0] || a.RequiredFeatures[1] != eh.RequiredFeatures[1] {
		return false
	}

	return true
}

func TestDecode(t *testing.T) {
	downloadTestOSMFile(t)

	f, err := os.Open(London)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	d := NewDecoder(f)
	d.SetBufferSize(1)

	header, err := d.Header()
	if err != nil {
		t.Fatal(err)
	}
	if checkHeader(header) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", eh, header)
	}

	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		t.Fatal(err)
	}

	var n *Node
	var w *Way
	var r *Relation
	var nc, wc, rc uint64
	var id string
	idsOrder := make([]string, 0, len(IDsExpectedOrder))
	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		} else {
			switch v := v.(type) {
			case *Node:
				nc++
				if v.ID == en.ID {
					n = v
				}
				id = fmt.Sprintf("node/%d", v.ID)
				if _, ok := IDs[id]; ok {
					idsOrder = append(idsOrder, id)
				}
			case *Way:
				wc++
				if v.ID == ew.ID {
					w = v
				}
				id = fmt.Sprintf("way/%d", v.ID)
				if _, ok := IDs[id]; ok {
					idsOrder = append(idsOrder, id)
				}
			case *Relation:
				rc++
				if v.ID == er.ID {
					r = v
				}
				id = fmt.Sprintf("relation/%d", v.ID)
				if _, ok := IDs[id]; ok {
					idsOrder = append(idsOrder, id)
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
	if enc != nc || ewc != wc || erc != rc {
		t.Errorf("\nExpected %7d nodes, %7d ways, %7d relations\nGot %7d nodes, %7d ways, %7d relations.",
			enc, ewc, erc, nc, wc, rc)
	}
	if !reflect.DeepEqual(IDsExpectedOrder, idsOrder) {
		t.Errorf("\nExpected: %v\nGot:      %v", IDsExpectedOrder, idsOrder)
	}
}

func TestDecodeConcurrent(t *testing.T) {
	downloadTestOSMFile(t)

	f, err := os.Open(London)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	d := NewDecoder(f)
	d.SetBufferSize(1)
	err = d.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		t.Fatal(err)
	}

	header, err := d.Header()
	if err != nil {
		t.Fatal(err)
	}
	if checkHeader(header) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", eh, header)
	}

	var n *Node
	var w *Way
	var r *Relation
	var nc, wc, rc uint64
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)

		go func() {
			for {
				if v, err := d.Decode(); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				} else {
					switch v := v.(type) {
					case *Node:
						atomic.AddUint64(&nc, 1)
						if v.ID == en.ID {
							n = v
						}
					case *Way:
						atomic.AddUint64(&wc, 1)
						if v.ID == ew.ID {
							w = v
						}
					case *Relation:
						atomic.AddUint64(&rc, 1)
						if v.ID == er.ID {
							r = v
						}
					default:
						t.Fatalf("unknown type %T", v)
					}
				}
			}

			wg.Done()
		}()
	}
	wg.Wait()

	if !reflect.DeepEqual(en, n) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", en, n)
	}
	if !reflect.DeepEqual(ew, w) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", ew, w)
	}
	if !reflect.DeepEqual(er, r) {
		t.Errorf("\nExpected: %#v\nActual:   %#v", er, r)
	}
	if enc != nc || ewc != wc || erc != rc {
		t.Errorf("\nExpected %7d nodes, %7d ways, %7d relations\nGot %7d nodes, %7d ways, %7d relations",
			enc, ewc, erc, nc, wc, rc)
	}
}

func BenchmarkDecode(b *testing.B) {
	file := os.Getenv("OSMPBF_BENCHMARK_FILE")
	if file == "" {
		file = London
	}
	f, err := os.Open(file)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	blobBufferSize, _ := strconv.Atoi(os.Getenv("OSMPBF_BENCHMARK_BUFFER"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)

		d := NewDecoder(f)
		if blobBufferSize > 0 {
			d.SetBufferSize(blobBufferSize)
		}
		err = d.Start(runtime.GOMAXPROCS(-1))
		if err != nil {
			b.Fatal(err)
		}

		var nc, wc, rc uint64
		start := time.Now()
		for {
			if v, err := d.Decode(); err == io.EOF {
				break
			} else if err != nil {
				b.Fatal(err)
			} else {
				switch v := v.(type) {
				case *Node:
					nc++
				case *Way:
					wc++
				case *Relation:
					rc++
				default:
					b.Fatalf("unknown type %T", v)
				}
			}
		}

		b.Logf("Done in %.3f seconds. Nodes: %d, Ways: %d, Relations: %d\n",
			time.Now().Sub(start).Seconds(), nc, wc, rc)
	}
}

func BenchmarkDecodeConcurrent(b *testing.B) {
	file := os.Getenv("OSMPBF_BENCHMARK_FILE")
	if file == "" {
		file = London
	}
	f, err := os.Open(file)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	blobBufferSize, _ := strconv.Atoi(os.Getenv("OSMPBF_BENCHMARK_BUFFER"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Seek(0, 0)

		d := NewDecoder(f)
		if blobBufferSize > 0 {
			d.SetBufferSize(blobBufferSize)
		}
		err = d.Start(runtime.GOMAXPROCS(-1))
		if err != nil {
			b.Fatal(err)
		}

		var nc, wc, rc uint64
		start := time.Now()
		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(1)

			go func() {
				for {
					if v, err := d.Decode(); err == io.EOF {
						break
					} else if err != nil {
						b.Fatal(err)
					} else {
						switch v := v.(type) {
						case *Node:
							atomic.AddUint64(&nc, 1)
						case *Way:
							atomic.AddUint64(&wc, 1)
						case *Relation:
							atomic.AddUint64(&rc, 1)
						default:
							b.Fatalf("unknown type %T", v)
						}
					}
				}

				wg.Done()
			}()
		}
		wg.Wait()

		b.Logf("Done in %.3f seconds. Nodes: %d, Ways: %d, Relations: %d\n",
			time.Now().Sub(start).Seconds(), nc, wc, rc)
	}
}
