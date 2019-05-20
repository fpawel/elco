package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
)

func Test_finsWriteFloat(t *testing.T) {
	n := math.Float32bits(float32(22.33))
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, n)
	b := buf.Bytes()

	words := []uint16{
		binary.LittleEndian.Uint16([]byte{b[0], b[1]}),
		binary.LittleEndian.Uint16([]byte{b[2], b[3]}),
	}

	fmt.Printf("% X - % X\n", b, words)
}
