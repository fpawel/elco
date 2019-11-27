package tst

import (
	"fmt"
	"os"
	"testing"
)

func TestPrintChipMemoryMap(_ *testing.T) {
	file, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	prn := func(i int, t float64) {
		_, _ = fmt.Fprintf(file, "%04X\t%v⁰C\n", i, t)
	}

	_, _ = file.WriteString("Фоновый ток\n")
	t := float64(0)
	for i := 0; i <= 0x00F8; i += 2 {
		prn(i, t)
		t--
	}
	t = 0
	for i := 0x0100; i <= 0x01F8; i += 2 {
		prn(i, t)
		t++
	}

	_, _ = file.WriteString("\n\n")
	_, _ = file.WriteString("Коэффициент чувствительности\n")
	t = float64(0)
	for i := 0x0400; i <= 0x04F8; i += 2 {
		prn(i, t)
		t--
	}
	t = 0
	for i := 0x0500; i <= 0x05F8; i += 2 {
		prn(i, t)
		t++
	}

	_, _ = file.WriteString("\n\n")
	_, _ = file.WriteString("Ввод\n")

	n := 0
	for i := 0x01FF; i < 0x0400; i += 4 + 4 + 4 {
		_, _ = fmt.Fprintf(file, "%04X\t[%d] температура\n", i, n)
		_, _ = fmt.Fprintf(file, "%04X\t[%d] ток\n", i+4, n)
		_, _ = fmt.Fprintf(file, "%04X\t[%d] Кч\n", i+4+4, n)
		n += 1
	}

}
