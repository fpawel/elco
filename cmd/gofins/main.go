package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fpawel/gofins/fins"
	"io/ioutil"
	"math"
)

func main() {

	flag.Parse()
	c := newConfig()
	fmt.Printf("%+v\n", c)
	cli, err := fins.NewClient(c.Client.Address(), c.Server.Address())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	cli.SetTimeoutMs(100)

	fmt.Println(finsReadFloat(cli, 2))

	setPoint, err := finsReadFloat(cli, 8)
	fmt.Println("setPoint", setPoint, err)

	fmt.Println(finsWriteFloat(cli, 8, -33.44))
	fmt.Println(finsReadFloat(cli, 8))
	fmt.Println(finsWriteFloat(cli, 8, setPoint))
	fmt.Println(finsReadFloat(cli, 8))

	fmt.Println(cli.ResetBit(fins.MemoryAreaWRBit, 0, 0))
	fmt.Println(cli.SetBit(fins.MemoryAreaWRBit, 0, 10))
}

func finsWriteFloat(finsClient *fins.Client, addr int, value float64) error {

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, math.Float32bits(float32(value)))
	b := buf.Bytes()
	words := []uint16{
		binary.LittleEndian.Uint16([]byte{b[0], b[1]}),
		binary.LittleEndian.Uint16([]byte{b[2], b[3]}),
	}

	fmt.Printf("% X - % X - %v\n", b, words, value)

	return finsClient.WriteWords(fins.MemoryAreaDMWord, uint16(addr), words)
}

func finsReadFloat(finsClient *fins.Client, addr int) (float64, error) {
	xs, err := finsClient.ReadBytes(fins.MemoryAreaDMWord, uint16(addr), 2)
	if err != nil {
		return 0, err
	}
	buf := bytes.NewBuffer([]byte{xs[1], xs[0], xs[3], xs[2]})
	var v float32
	if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
		return 0, err
	}
	return float64(v), nil
}

type config struct {
	Client, Server finsConfig
}

type finsConfig struct {
	IP                  string
	Port                int
	Network, Node, Unit byte
}

func newConfig() config {
	b, err := ioutil.ReadFile("config.json")
	var x config
	if err == nil {
		err = json.Unmarshal(b, &x)
	}
	if err != nil {
		fmt.Println(err)
		x = config{
			Server: finsConfig{
				IP:   "192.168.250.1",
				Port: 9600,
				Node: 1,
			},
			Client: finsConfig{
				IP:   "192.168.250.3",
				Port: 9600,
				Node: 254,
			},
		}

		b, err := json.MarshalIndent(x, "", "    ")
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile("config.json", b, 0666); err != nil {
			panic(err)
		}
	}
	return x
}
func (x finsConfig) Address() fins.Address {
	return fins.NewAddress(x.IP, x.Port, x.Network, x.Node, x.Unit)
}
