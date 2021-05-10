all:
	# sync version with ../go.mod
	go install -v google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0

	rm *.go
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go_opt=Mfileformat.proto=github.com/qedus/osmpbf/OSMPBF \
		--go_opt=Mosmformat.proto=github.com/qedus/osmpbf/OSMPBF \
		*.proto
