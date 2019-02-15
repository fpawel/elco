package main

// программа для простой проверки modbus
// -port=COM2 -addr=17 -data="00 02 00 02" -cmd=3 -rand=0 -cycle=5

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/fpawel/elco/pkg/serial-comm/comm"
	"github.com/fpawel/elco/pkg/serial-comm/comport"
	"github.com/fpawel/elco/pkg/serial-comm/modbus"
	"github.com/tarm/serial"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

func main() {
	var (
		comportName,
		strData string
		repeatCount,
		timeout,
		addr, randBytesCount, cmd int
		help bool
		req  modbus.Req
		boud int
	)
	flag.BoolVar(&help, "?", false, "использование программы")
	flag.StringVar(&comportName, "port", "COM1", "имя компорта")
	flag.IntVar(&repeatCount, "cycle", 1, "количество повторений")
	flag.IntVar(&addr, "addr", 1, "адрес модбас")
	flag.IntVar(&timeout, "timeout", 1000, "таймаут считывания, мс")
	flag.IntVar(&randBytesCount, "rand", 0, "количество сгенерированных байт данных со случайным значением")
	flag.IntVar(&cmd, "cmd", 0, "код команды")
	flag.IntVar(&boud, "boud", 9600, "скорость передачи, бод")
	flag.StringVar(&strData, "data", "", "основные данные")

	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	if cmd == 0 {
		fmt.Println("не задан код команды!")
		return
	}

	req.Addr = modbus.Addr(addr)
	req.ProtoCmd = modbus.ProtoCmd(cmd)

	if len(strData) > 0 {
		r := regexp.MustCompile(`[\dABCDEFabcdef]{2}`)
		tokens := r.FindAllString(strData, -1)
		if len(tokens) == 0 {
			fmt.Printf("%q: не верный формат", strData)
			return
		}
		for i, str := range r.FindAllString(strData, -1) {
			v, err := strconv.ParseInt(str, 16, 9)
			if err != nil {
				fmt.Printf("data[%d]: %v", i, err)
				return
			}
			req.Data = append(req.Data, byte(v))
		}
	}
	if randBytesCount > 0 {
		rndSrc := rand.NewSource(time.Now().UnixNano())
		rnd := rand.New(rndSrc)
		xs := make([]byte, randBytesCount)
		rnd.Read(xs)
		req.Data = append(req.Data, xs...)
	}

	portConfig := comport.Config{
		Serial: serial.Config{
			Name:        comportName,
			Baud:        boud,
			ReadTimeout: time.Millisecond,
		},
		Uart: comm.Config{
			MaxAttemptsRead: 1,
			ReadTimeout:     time.Millisecond * time.Duration(timeout),
			ReadByteTimeout: 50 * time.Millisecond,
		},
	}

	port := comport.NewPort(context.Background())
	if err := port.Open(portConfig); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("REQUEST:\n" + hex.Dump(req.Bytes()) + "\n")

	for i := 0; repeatCount == 0 || i < repeatCount; i++ {
		t := time.Now()
		bytes, err := port.GetResponse(req.Bytes())
		if err != nil {
			fmt.Printf("[%d] %v %v\n", i+1, err, time.Since(t))
		} else {
			fmt.Printf("[%d] %v\n%v", i+1, time.Since(t), hex.Dump(bytes))
		}
	}
	fmt.Println("close port:", port.Close())

}
