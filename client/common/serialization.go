package common

import (
	"bytes"
	"errors"
)

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

func SBDSerializeArray(items [][]string) []byte {
	data := make([]byte, 0)

	for _, item := range items {
		data = append(data, SBDSerialize(item)...)
		data = append(data, []byte{0}...)
	}

	// Remove the last 0 byte
	if len(data) > 0 {
		data = data[:len(data)-1]
	}

	return data
}

func SBDDeserialize(data []byte) []string {
	parts := bytes.Split(data, []byte{0})
	strings := make([]string, len(parts))

	for i, part := range parts {
		strings[i] = string(part)
	}

	return strings
}

func SBDDeserializeArray(data []byte, expected_parts int) ([][]string, error) {
	parts := SBDDeserialize(data)

	if len(parts) == 0 {
		return make([][]string, 0), nil
	}

	if len(parts) % expected_parts != 0 {
		return nil, errors.New("invalid number of parts")
	}

	items_qty := len(parts) / expected_parts
	items := make([][]string, items_qty)

	for i := 0; i < items_qty; i ++ {
		item := make([]string, expected_parts)

		for j := 0; j < expected_parts; j++ {
			item[j] = parts[(i * expected_parts) + j]
		}

		items[i] = item
	}

	return items, nil
}