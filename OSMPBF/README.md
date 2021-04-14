# Proto Files

*.proto files were downloaded from https://github.com/scrosby/OSM-binary/tree/master/src and changed in following ways:

## Changes
### StringTable
- **File**: osmformat.proto
- **Reason**: To eliminate continuous conversions from `[]byte` to `string`
- **Old Code**:
```protobuf
message StringTable {
   repeated bytes s = 1;
}
```
- **New Code**:

```protobuf
message StringTable {
   repeated string s = 1;
}
```
- **Comptatibility**: This changes is expected to be fully compatible with all PBF files.

### Optimization Goal
- **File**: osmformat.proto, fileformat.proto
- **Reason**: Better performance
- **Added Code**:

```protobuf
option optimize_for = LITE_RUNTIME;
```
- **Comptatibility**: Descriptors or Reflection are not available.


### StringTable
- **File**: osmformat.proto, fileformat.proto
- **Reason**: Required for generation
- **Old Code**:
```protobuf
option java_package = "crosby.binary";
```
- **New Code**:

```protobuf
option go_package = "OSMPBF";
```
- **Comptatibility**: This changes is expected to be fully compatible with all PBF files.

