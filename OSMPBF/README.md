# Proto Files

`*.proto` files were downloaded from https://github.com/scrosby/OSM-binary/tree/master/src and changed in following ways:

## Changes

### StringTable

- **File**: `osmformat.proto`
- **Reason**: To eliminate continuous conversions from `[]byte` to `string`
- **Old code**:

```protobuf
message StringTable {
   repeated bytes s = 1;
}
```

- **New code**:

```protobuf
message StringTable {
   repeated string s = 1;
}
```

- **Comptatibility**: This change is expected to be fully compatible with all PBF files.
