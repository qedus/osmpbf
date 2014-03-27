package osmpbf

func extractTags(stringTable []string, keyIDs, valueIDs []uint32) map[string]string {
	tags := make(map[string]string, len(keyIDs))
	for index, keyID := range keyIDs {
		key := stringTable[keyID]
		val := stringTable[valueIDs[index]]
		tags[key] = val
	}
	return tags
}
