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
	"io/ioutil"
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

// A Decoder reads and decodes OpenStreetMap PBF data from an input stream.
type Decoder struct {
	r io.Reader

	objectQueue []interface{}
	objectIndex int
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r, make([]interface{}, 0), 0}
}

// Decode reads the next object from the input stream and returns either a
// Node, Way or Relation struct representing the underlying OpenStreetMap PBF
// data.
//
// The end of the input stream is reported by an io.EOF error.
func (dec *Decoder) Decode() (interface{}, error) {
	if dec.objectIndex >= len(dec.objectQueue) {
		dec.objectQueue = dec.objectQueue[:0]
		dec.objectIndex = 0
		if err := dec.readNextPrimitiveBlock(); err != nil {
			return nil, err
		}
	}

	dec.objectIndex++
	return dec.objectQueue[dec.objectIndex-1], nil
}

func (dec *Decoder) readNextPrimitiveBlock() error {
	for {
		blobHeaderSize, err := dec.readBlobHeaderSize()
		if err != nil {
			return err
		}

		blobHeader, err := dec.readBlobHeader(blobHeaderSize)
		if err != nil {
			return err
		}

		blob, err := dec.readBlob(blobHeader)
		if err != nil {
			return err
		}

		switch blobHeader.GetType() {
		case "OSMHeader":
			if err := dec.readOSMHeader(blob); err != nil {
				return err
			}
		case "OSMData":
			return dec.readOSMData(blob)
		default:
			// Skip over unknown type
		}
	}
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

	blobHeader := &OSMPBF.BlobHeader{}
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

	blob := &OSMPBF.Blob{}
	if err := proto.Unmarshal(buf, blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func (dec *Decoder) readOSMHeader(blob *OSMPBF.Blob) error {
	data, err := getData(blob)
	if err != nil {
		return err
	}
	headerBlock := &OSMPBF.HeaderBlock{}
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return fmt.Errorf("parser does not have %s capability",
				feature)
		}
	}
	return nil
}

func (dec *Decoder) readOSMData(blob *OSMPBF.Blob) error {
	data, err := getData(blob)
	if err != nil {
		return err
	}

	primitiveBlock := &OSMPBF.PrimitiveBlock{}
	if err := proto.Unmarshal(data, primitiveBlock); err != nil {
		return err
	}
	dec.parsePrimitiveBlock(primitiveBlock)
	return nil
}

func getData(blob *OSMPBF.Blob) ([]byte, error) {
	switch {
	case blob.Raw != nil:
		return blob.GetRaw(), nil
	case blob.ZlibData != nil:
		compressedData := bytes.NewBuffer(blob.GetZlibData())
		r, err := zlib.NewReader(compressedData)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		if len(data) != int(blob.GetRawSize()) {
			return nil, fmt.Errorf(
				"raw blob data size %d but expected %d",
				len(data), blob.GetRawSize())
		}
		return data, nil
	}
	return nil, errors.New("unknown blob data")
}

func (dec *Decoder) parsePrimitiveBlock(pb *OSMPBF.PrimitiveBlock) {
	primitiveGroups := pb.GetPrimitivegroup()
	for _, pg := range primitiveGroups {
		dec.parsePrimitiveGroup(pb, pg)
	}
}

func (dec *Decoder) parsePrimitiveGroup(pb *OSMPBF.PrimitiveBlock, pg *OSMPBF.PrimitiveGroup) {
	dec.parseNodes(pb, pg.GetNodes())
	dec.parseDenseNodes(pb, pg.GetDense())
	dec.parseWays(pb, pg.GetWays())
	dec.parseRelations(pb, pg.GetRelations())
}

func (dec *Decoder) parseNodes(pb *OSMPBF.PrimitiveBlock, nodes []*OSMPBF.Node) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())
	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()

	for _, node := range nodes {
		id := node.GetId()
		lat := node.GetLat()
		lon := node.GetLon()

		latitude := 1e-9 * float64((latOffset + (granularity * lat)))
		longitude := 1e-9 * float64((lonOffset + (granularity * lon)))

		tags := extractTags(st, node.GetKeys(), node.GetVals())
		dec.objectQueue = append(dec.objectQueue,
			&Node{id, latitude, longitude, tags})
		panic("Please test this first")
	}
}

func (dec *Decoder) parseDenseNodes(pb *OSMPBF.PrimitiveBlock, dn *OSMPBF.DenseNodes) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())
	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()
	ids := dn.GetId()
	lats := dn.GetLat()
	lons := dn.GetLon()
	tu := tagUnpacker{st, dn.GetKeysVals(), 0}
	id, lat, lon := int64(0), int64(0), int64(0)
	for index := range ids {
		id = ids[index] + id
		lat = lats[index] + lat
		lon = lons[index] + lon
		latitude := 1e-9 * float64((latOffset + (granularity * lat)))
		longitude := 1e-9 * float64((lonOffset + (granularity * lon)))
		tags := tu.next()
		dec.objectQueue = append(dec.objectQueue,
			&Node{id, latitude, longitude, tags})
	}
}

type tagUnpacker struct {
	stringTable [][]byte
	keysVals    []int32
	index       int
}

func (tu *tagUnpacker) next() map[string]string {
	tags := make(map[string]string)
	for ; tu.index < len(tu.keysVals); tu.index++ {

		keyID := tu.keysVals[tu.index]
		tu.index++
		if keyID == 0 {
			break
		}
		valID := tu.keysVals[tu.index]
		key := string(tu.stringTable[keyID])
		val := string(tu.stringTable[valID])
		tags[key] = val
	}
	return tags
}

func extractTags(stringTable [][]byte, keyIDs, valueIDs []uint32) map[string]string {
	tags := make(map[string]string)
	for index := range keyIDs {
		key := string(stringTable[keyIDs[index]])
		val := string(stringTable[valueIDs[index]])
		tags[key] = val
	}
	return tags
}

func (dec *Decoder) parseWays(pb *OSMPBF.PrimitiveBlock, ways []*OSMPBF.Way) {
	st := pb.GetStringtable().GetS()
	for _, way := range ways {
		id := way.GetId()
		tags := extractTags(st, way.GetKeys(), way.GetVals())

		// NodeIDs
		refs := way.GetRefs()
		nodeID := int64(0)
		nodeIDs := make([]int64, 0, len(refs))
		for index := range refs {
			nodeID = refs[index] + nodeID
			nodeIDs = append(nodeIDs, nodeID)
		}
		dec.objectQueue = append(dec.objectQueue,
			&Way{id, tags, nodeIDs})
	}
}

func extractMembers(stringTable [][]byte, rel *OSMPBF.Relation) []Member {
	memIDs := rel.GetMemids()
	types := rel.GetTypes()
	roleIDs := rel.GetRolesSid()

	memID := int64(0)
	members := make([]Member, 0, len(memIDs))
	for index := range memIDs {
		memID = memIDs[index] + memID
		memType := NodeType
		switch types[index] {
		case OSMPBF.Relation_WAY:
			memType = WayType
		case OSMPBF.Relation_RELATION:
			memType = RelationType
		}
		role := string(stringTable[roleIDs[index]])
		members = append(members, Member{memID, memType, role})
	}
	return members
}

func (dec *Decoder) parseRelations(pb *OSMPBF.PrimitiveBlock, relations []*OSMPBF.Relation) {
	st := pb.GetStringtable().GetS()
	for _, rel := range relations {
		id := rel.GetId()
		tags := extractTags(st, rel.GetKeys(), rel.GetVals())
		members := extractMembers(st, rel)

		dec.objectQueue = append(dec.objectQueue,
			&Relation{id, tags, members})
	}
}
