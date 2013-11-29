package pbf_test

import (
	"fmt"
	"github.com/qedus/pbf"
	"io"
	"os"
	"testing"
)

func TestDecoder(t *testing.T) {
	f, err := os.Open("greater_london.osm.pbf")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer f.Close()
	d := pbf.NewDecoder(f)

	for {
		if v, err := d.Decode(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err.Error())
		}
	}
}
