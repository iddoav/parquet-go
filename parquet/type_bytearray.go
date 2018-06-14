package parquet

import (
	"encoding/binary"
	"fmt"
)

type byteArrayPlainDecoder struct {
	// length > 0 for FIXED_BYTE_ARRAY type
	length int

	data []byte

	pos int
}

func (d *byteArrayPlainDecoder) init(data []byte, count int) error {
	d.data = data
	d.pos = 0
	return nil
}

func (d *byteArrayPlainDecoder) next() (value []byte, err error) {
	size := d.length
	if d.length == 0 {
		if d.pos > len(d.data)-4 {
			return nil, fmt.Errorf("bytearray/plain: no more data")
		}
		size = int(binary.LittleEndian.Uint32(d.data[d.pos:])) // TODO: think about int overflow here
		d.pos += 4
	}
	if d.pos > len(d.data)-size {
		return nil, fmt.Errorf("bytearray/plain: not enough data")
	}
	// TODO: configure copy or not
	value = make([]byte, size)
	copy(value, d.data[d.pos:d.pos+size])
	d.pos += size
	return value, err
}

func (d *byteArrayPlainDecoder) decode(slice interface{}) (n int, err error) {
	// TODO: support string
	switch buf := slice.(type) {
	case [][]byte:
		return d.decodeByteSlice(buf)
	case []interface{}:
		return d.decodeE(buf)
	default:
		panic("invalid argument")
	}
}

func (d *byteArrayPlainDecoder) decodeByteSlice(buf [][]byte) (n int, err error) {
	i := 0
	for i < len(buf) && d.pos < len(d.data) {
		buf[i], err = d.next()
		if err != nil {
			break
		}
		i++
	}
	if i == 0 {
		err = fmt.Errorf("bytearray/plain: no more data")
	}
	return i, err
}

func (d *byteArrayPlainDecoder) decodeE(buf []interface{}) (n int, err error) {
	b := make([][]byte, len(buf), len(buf))
	n, err = d.decodeByteSlice(b)
	for i := 0; i < n; i++ {
		buf[i] = b[i]
	}
	return n, err
}

type byteArrayDictDecoder struct {
	dictDecoder

	values [][]byte
}

func (d *byteArrayDictDecoder) initValues(dictData []byte, count int) error {
	d.numValues = count
	d.values = make([][]byte, count, count)
	return d.dictDecoder.initValues(d.values, dictData)
}

func (d *byteArrayDictDecoder) decode(slice interface{}) (n int, err error) {
	// TODO: support string
	switch buf := slice.(type) {
	case [][]byte:
		return d.decodeByteSlice(buf)
	case []interface{}:
		return d.decodeE(buf)
	default:
		panic("invalid argument")
	}
}

func (d *byteArrayDictDecoder) decodeByteSlice(buf [][]byte) (n int, err error) {
	keys, err := d.decodeKeys(len(buf))
	if err != nil {
		return 0, err
	}
	for i, k := range keys {
		buf[i] = d.values[k]
	}
	return len(keys), nil
}

func (d *byteArrayDictDecoder) decodeE(buf []interface{}) (n int, err error) {
	b := make([][]byte, len(buf), len(buf))
	n, err = d.decodeByteSlice(b)
	for i := 0; i < n; i++ {
		buf[i] = b[i]
	}
	return n, err
}
