// Package osmpbf decodes OpenStreetMap pbf files.
// Use this package by creating a NewDecoder and passing it a PBF file. Use
// Decode to return Node, Way and Relation structs.
package osmpbf

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/qedus/osmpbf/OSMPBF"
	"io"
)

const (
	maxBlobHeaderSize = 64 * 1024
	maxBlobSize       = 32 * 1024 * 1024
)

var (
	parseCapabilities = map[string]bool{
		"OsmSchema-V0.6": true,
		"DenseNodes":     true,
	}
)

type Node struct {
	ID   int64
	Lat  float64
	Lon  float64
	Tags map[string]string

	// TODO: Add DenseInfo
}

type Way struct {
	ID      int64
	Tags    map[string]string
	NodeIDs []int64

	// TODO: Add Info
}

type Relation struct {
	ID      int64
	Tags    map[string]string
	Members []Member

	// TODO: Add Info
	// TODO: Add roles_sid
}

type MemberType int

const (
	NodeType MemberType = iota
	WayType
	RelationType
)

type Member struct {
	ID   int64
	Type MemberType
	Role string
}

type pair struct {
	i interface{}
	e error
}

// A Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	r          io.Reader
	serializer chan pair

	// for data decoders
	inputs  []chan<- pair
	outputs []<-chan pair
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:          r,
		serializer: make(chan pair, 8000), // typical PrimitiveBlock contains 8k OSM entities
	}
}

// Start decoding process using n goroutines.
func (dec *Decoder) Start(n int) error {
	// read OSMHeader
	blobHeader, blob, err := dec.readFileBlock()
	if err == nil {
		if blobHeader.GetType() == "OSMHeader" {
			err = decodeOSMHeader(blob)
		} else {
			err = fmt.Errorf("unexpected first fileblock of type %s", blobHeader.GetType())
		}
	}
	if err != nil {
		return err
	}

	// start data decoders
	for i := 0; i < n; i++ {
		input := make(chan pair)
		output := make(chan pair)
		go func() {
			dd := new(dataDecoder)
			for p := range input {
				if p.e == nil {
					// send decoded objects or decoding error
					objects, err := dd.Decode(p.i.(*OSMPBF.Blob))
					output <- pair{objects, err}
				} else {
					// send input error as is
					output <- pair{nil, p.e}
				}
			}
			close(output)
		}()

		dec.inputs = append(dec.inputs, input)
		dec.outputs = append(dec.outputs, output)
	}

	// start reading OSMData
	go func() {
		var inputIndex int
		for {
			input := dec.inputs[inputIndex]
			inputIndex = (inputIndex + 1) % n

			blobHeader, blob, err = dec.readFileBlock()
			if err == nil && blobHeader.GetType() != "OSMData" {
				err = fmt.Errorf("unexpected fileblock of type %s", blobHeader.GetType())
			}
			if err == nil {
				// send blob for decoding
				input <- pair{blob, nil}
			} else {
				// send input error as is
				input <- pair{nil, err}
				for _, input := range dec.inputs {
					close(input)
				}
				return
			}
		}
	}()

	go func() {
		var outputIndex int
		for {
			output := dec.outputs[outputIndex]
			outputIndex = (outputIndex + 1) % n

			p := <-output
			if p.i != nil {
				// send decoded objects one by one
				for _, o := range p.i.([]interface{}) {
					dec.serializer <- pair{o, nil}
				}
			}
			if p.e != nil {
				// send input or decoding error
				dec.serializer <- pair{nil, p.e}
				close(dec.serializer)
				return
			}
		}
	}()

	return nil
}

// Decode reads the next object from the input stream and returns either a
// Node, Way or Relation struct representing the underlying OpenStreetMap PBF
// data, or error encountered. The end of the input stream is reported by an io.EOF error.
//
// Decode is safe for parallel execution. Only first error encountered will be returned,
// subsequent invocations will return io.EOF.
func (dec *Decoder) Decode() (interface{}, error) {
	p, ok := <-dec.serializer
	if !ok {
		return nil, io.EOF
	}
	return p.i, p.e
}

func (dec *Decoder) readFileBlock() (*OSMPBF.BlobHeader, *OSMPBF.Blob, error) {
	blobHeaderSize, err := dec.readBlobHeaderSize()
	if err != nil {
		return nil, nil, err
	}

	blobHeader, err := dec.readBlobHeader(blobHeaderSize)
	if err != nil {
		return nil, nil, err
	}

	blob, err := dec.readBlob(blobHeader)
	if err != nil {
		return nil, nil, err
	}

	return blobHeader, blob, err
}

func (dec *Decoder) readBlobHeaderSize() (uint32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return 0, err
	}
	size := binary.BigEndian.Uint32(buf)

	if size >= maxBlobHeaderSize {
		return 0, errors.New("BlobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *Decoder) readBlobHeader(size uint32) (*OSMPBF.BlobHeader, error) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blobHeader := new(OSMPBF.BlobHeader)
	if err := proto.Unmarshal(buf, blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= maxBlobSize {
		return nil, errors.New("Blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *Decoder) readBlob(blobHeader *OSMPBF.BlobHeader) (*OSMPBF.Blob, error) {
	buf := make([]byte, blobHeader.GetDatasize())
	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return nil, err
	}

	blob := new(OSMPBF.Blob)
	if err := proto.Unmarshal(buf, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func getData(blob *OSMPBF.Blob) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil

	case blob.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(make([]byte, 0, blob.GetRawSize()+bytes.MinRead))
		_, err = buf.ReadFrom(r)
		if err != nil {
			return nil, err
		}
		if buf.Len() != int(blob.GetRawSize()) {
			err = fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
			return nil, err
		}
		return buf.Bytes(), nil

	default:
		return nil, errors.New("unknown blob data")
	}
}

func decodeOSMHeader(blob *OSMPBF.Blob) error {
	data, err := getData(blob)
	if err != nil {
		return err
	}

	headerBlock := new(OSMPBF.HeaderBlock)
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	return nil
}
