package main

import (
	"github.com/fpawel/comm/modbus"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"bytes"
	"encoding/binary"

	"fmt"
	"strconv"
	"strings"
)

var mw *walk.MainWindow
var neIn, neFloat, neBCD, leCRC16In, leCRC16Out *walk.LineEdit
var editMode bool

func onCRC16() {
	h, l := modbus.CRC16(parseBytes(leCRC16In.Text()))
	leCRC16Out.SetText(fmt.Sprintf("% X", []byte{h, l}))
}

func parseBytes(str string) (b []byte) {
	for _, s := range strings.Fields(str) {
		n, err := strconv.ParseUint(s, 16, 8)
		if n > 255 || err != nil {
			return
		}
		b = append(b, byte(n))
	}
	return
}

func updateBCD(value float64) {
	neBCD.SetText(fmt.Sprintf("% X", modbus.BCD6(value)))
}

func updateFloat(value float64) {
	buf := bytes.NewBuffer(nil)

	binary.Write(buf, binary.LittleEndian, float32(value))
	neFloat.SetText(fmt.Sprintf("% X", buf.Bytes()))

}

func neInChanged() {
	if editMode {
		return
	}
	editMode = true
	defer func() {
		editMode = false
	}()

	value, err := strconv.ParseFloat(neIn.Text(), 32)
	neFloat.SetText("?")
	neBCD.SetText("?")
	if err == nil {
		updateBCD(value)
		updateFloat(value)
	}
}

func neFloatChanged() {
	if editMode {
		return
	}
	editMode = true
	defer func() {
		editMode = false
	}()

	neIn.SetText("?")
	neBCD.SetText("?")

	b := parseBytes(neFloat.Text())
	buf := bytes.NewReader(b)
	var v float32
	err := binary.Read(buf, binary.LittleEndian, &v)
	if err == nil {
		neIn.SetText(fmt.Sprintf("%g", v))
		updateBCD(float64(v))
	}

}

func neBCDChanged() {
	if editMode {
		return
	}
	editMode = true
	defer func() {
		editMode = false
	}()

	neIn.SetText("?")
	neFloat.SetText("?")

	b := parseBytes(neBCD.Text())
	if len(b) != 4 {
		return
	}
	v, ok := modbus.ParseBCD6(parseBytes(neBCD.Text()))
	if ok {
		neIn.SetText(fmt.Sprintf("%g", v))
		updateFloat(float64(v))
	}

}

func main() {

	err := MainWindow{
		Title:    "Калькулятор Аналитприбора",
		Size:     Size{600, 450},
		Layout:   Grid{},
		AssignTo: &mw,
		Font:     Font{PointSize: 12, Family: "Segoe UI"},

		Children: []Widget{
			ScrollView{
				Layout: VBox{},
				Children: []Widget{
					Label{
						Text: "Число:",
					},
					LineEdit{
						AssignTo:      &neIn,
						OnTextChanged: neInChanged,
					},
					Label{
						Text: "float:",
					},
					LineEdit{
						AssignTo:      &neFloat,
						OnTextChanged: neFloatChanged,
					},
					Label{
						Text: "bcd:",
					},
					LineEdit{
						AssignTo:      &neBCD,
						OnTextChanged: neBCDChanged,
					},
					Label{
						Text: "CRC16:",
					},
					ScrollView{
						Layout: HBox{},
						Children: []Widget{

							LineEdit{
								AssignTo:      &leCRC16In,
								OnTextChanged: onCRC16,
							},
							LineEdit{
								AssignTo: &leCRC16Out,
								MaxSize:  Size{100, 0},
							},
						},
					},
				},
			},
		},
	}.Create()
	if err != nil {
		panic(err)
	}
	mw.Run()

	return
}
