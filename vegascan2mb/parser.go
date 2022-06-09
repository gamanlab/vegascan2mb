package main

import (
	"encoding/binary"
	"encoding/csv"
	"strconv"
	"strings"
)

func ParseCSVString(input string) ([][]string, error) {
	strReader := strings.NewReader(input)
	csvReader := csv.NewReader(strReader)
	csvReader.Comma = ';'
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true
	csvReader.Comment = '#'

	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func CsvToUint32(input string) ([]uint32, error) {
	stringData, err := ParseCSVString(input)
	if err != nil {
		return nil, err
	}

	values := make([]uint32, 0)

	for _, strData := range stringData {
		value64, err := strconv.ParseUint(strData[2], 10, 0)
		if err != nil {
			return nil, err
		}

		value32 := uint32(value64)
		values = append(values, value32)

	}

	return values, nil
}

// Uint32ToBytesBE converts an array of uint16s to a big endian array of bytes
func Uint32ToBytesBE(values []uint32) []byte {
	bytes := make([]byte, len(values)*4)

	for i, value := range values {
		binary.BigEndian.PutUint32(bytes[i*4:(i+1)*4], value)
	}
	return bytes
}

// Uint32ToBytesLE converts an array of uint16s to a big endian array of bytes
func Uint32ToBytesLE(values []uint32) []byte {
	bytes := make([]byte, len(values)*4)

	for i, value := range values {
		binary.LittleEndian.PutUint32(bytes[i*4:(i+1)*4], value)
	}
	return bytes
}

func CsvStringToBytes(input string, LE bool) ([]byte, error) {
	vals, err := CsvToUint32(input)
	if err != nil {
		return nil, err
	}

	var byteVals []byte
	if LE {
		byteVals = Uint32ToBytesLE(vals)
	} else {
		byteVals = Uint32ToBytesBE(vals)
	}

	return byteVals, nil
}
