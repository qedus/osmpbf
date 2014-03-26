package osmpbf

func extractTags(stringTable [][]byte, keyIDs, valueIDs []uint32) map[string]string {
	tags := make(map[string]string, len(keyIDs))
	for index, keyID := range keyIDs {
		key := string(stringTable[keyID])
		val := string(stringTable[valueIDs[index]])
		tags[key] = val
	}
	return tags
}
