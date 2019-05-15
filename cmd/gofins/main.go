package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/l1va/gofins/fins"
	"io/ioutil"
)

func main() {

	area := flag.Int("area", int(fins.MemoryAreaDMWord), "область")
	addr := flag.Int("addr", 8, "адрес")
	count := flag.Int("count", 2, "количество байт")
	timeout := flag.Int("timeout", 100, "таймаут, мс")

	flag.Parse()

	c := newConfig()
	fmt.Printf("%+v\n", c)
	cli, err := fins.NewClient(c.Client.Address(), c.Server.Address())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	cli.SetTimeoutMs(uint(*timeout))

	xs, err := cli.ReadBytes(byte(*area), uint16(*addr), uint16(*count))
	if err != nil {
		panic(err)
	}
	fmt.Printf("% X", xs)
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
