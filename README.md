pbf
===

OpenStreetMap PBF file format parser in Go Lang.

Instructions:
1) Clone this repository.
2) Make sure your GOPATH is set to this repository.
3) Install Google Protocol Buffer library.
4) Get Go Lang Protocol Buffer code with:
    go get -u code.google.com/p/goprotobuf/{proto,protoc-gen-go}
5) Put /bin on your PATH as protoc-gen-go is needed to be found there.
6) In this library's package build the fileformat.pb.go and osmformat.pb.go files with:
    protoc --go_out=. *.proto
Note that protoc is in the bin directory of wherever you installed Google Protocol Buffers.
