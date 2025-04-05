package common

import "bytes"

// SBDSerialization (Sequential Binary Data Serialization) provides
// binary serialization and deserialization for a list of ordered strings.

// Serialize a sequence of strings into a sequential binary strings
// bytes representation.
func SBDSerialize(parts []string) []byte {
	data := make([]byte, 0)

	for _, part := range parts {
		data = append(data, []byte(part)...)
		data = append(data, []byte{0}...)
	}

	// Remove the last 0 byte
	if len(data) > 0 {
		data = data[:len(data)-1]
	}

	return data
}

func SBDDeserialize(data []byte) []string {
	// Go's Split method will return a slice with a single empty string
	// even if the input is empty, to avoid that an early return
	// is used to return an empty slice (length zero).
	if len(data) == 0 {
		return make([]string, 0)
	}

	parts := bytes.Split(data, []byte{0})
	strings := make([]string, len(parts))

	for i, part := range parts {
		strings[i] = string(part)
	}

	return strings
}