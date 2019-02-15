package comport

import (
	"fmt"
	"testing"
	"time"
)

func TestGetAvailablePorts(t *testing.T) {
	ports, err := AvailablePorts()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Ports: %+q\n", ports)

	ch := make(chan bool)
	go func() {
		for {
			<-ch
			ports, err := AvailablePorts()
			if err != nil {
				t.Error(err)
			}
			fmt.Printf("%v: %+q\n", time.Now(), ports)
		}

	}()
	for {
		NotifyAvailablePortsChange(ch)
	}

}

func TestNotify(t *testing.T) {
	notifyWindow.NotifyJson(0, struct {
		Com   string
		Error bool
		Msg   string
	}{"COM1", true, "01 02 03 04 05"})
	fmt.Println("ok")
}
