package bencode

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"strconv"
)

type Decoder struct {
	data []byte
	pos  int
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data: data,
		pos:  0,
	}
}

func Unmarshal(data []byte) (any, error) {
	d := NewDecoder(data)
	return d.Decode()
}

func (d *Decoder) Decode() (any, error) {
	if d.eof() {
		return nil, fmt.Errorf("unexpected EOF at position %d", d.pos)
	}

	b := d.peek()

	switch {
	case isDigit(b):
		return d.decodeString()

	case b == 'i':
		return d.decodeInt()

	case b == 'l':
		return d.decodeList()

	case b == 'd':
		return d.decodeDict()

	default:
		return nil, fmt.Errorf("enexpected byte '%c' at position %d", b, d.pos)
	}
}

func (d *Decoder) peek() byte {
	return d.data[d.pos]
}

func (d *Decoder) eof() bool {
	return d.pos >= len(d.data)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func (d *Decoder) decodeString() (string, error) {
	colonIdx := bytes.IndexByte(d.data[d.pos:], ':')
	if colonIdx == -1 {
		return "", fmt.Errorf("missing ':' for string length at position %d", d.pos)
	}

	absColonPos := d.pos + colonIdx

	length, err := strconv.Atoi(string(d.data[d.pos:absColonPos]))
	if err != nil {
		return "", fmt.Errorf("invalid string length: %w", err)
	}

	d.pos = absColonPos + 1

	if d.pos+length > len(d.data) {
		return "", fmt.Errorf("string bounds exceeded at position %d", d.pos)
	}

	value := string(d.data[d.pos : d.pos+length])
	d.pos += length

	return value, nil
}

func (d *Decoder) decodeInt() (int, error) {
	d.pos++

	endIdx := bytes.IndexByte(d.data[d.pos:], 'e')
	if endIdx == -1 {
		return 0, fmt.Errorf("missing 'e' for integer at position %d", d.pos)
	}

	absEndPos := d.pos + endIdx

	numStr := string(d.data[d.pos:absEndPos])

	if len(numStr) > 1 && (numStr[0] == '0' || numStr[:2] == "-0") {
		return 0, fmt.Errorf("number cannot start with 0")
	}

	val, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid integer format: %w", err)
	}

	d.pos = absEndPos + 1

	return val, nil
}

func (d *Decoder) decodeList() ([]any, error) {
	d.pos++

	var list []any

	for !d.eof() {
		if d.peek() == 'e' {
			d.pos++
			return list, nil
		}

		val, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode list element: %w", err)
		}

		list = append(list, val)
	}

	return nil, fmt.Errorf("unexpected EOF: missing 'e' to close list")
}

func (d *Decoder) decodeDict() (map[string]any, error) {
	d.pos++

	dict := make(map[string]any)

	for !d.eof() {
		if d.peek() == 'e' {
			d.pos++
			return dict, nil
		}

		keyAny, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode dictionary key: %w", err)
		}

		strKey, ok := keyAny.(string)
		if !ok {
			return nil, fmt.Errorf("dictionary key must be string, got %T", keyAny)
		}

		if strKey == "info" {
			rawStart := d.pos

			value, err := d.Decode()
			if err != nil {
				return nil, fmt.Errorf("failed to decode dict value for key %q: %w", strKey, err)
			}

			rawEnd := d.pos
			rawInfoBytes := d.data[rawStart:rawEnd]

			infoHash := sha1.Sum(rawInfoBytes)

			dict[strKey] = value
			dict["_info_hash"] = infoHash

			continue
		}

		value, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode dictionary value for key %q: %w", strKey, err)
		}

		dict[strKey] = value
	}

	return nil, fmt.Errorf("unexpected EOF: missing 'e' to close dictionary")
}
