package comport

import (
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/winapp"
	"golang.org/x/sys/windows/registry"
	"log"
	"syscall"
)

func NotifyAvailablePortsChange(notifyCh chan<- bool) (err error) {
	var regKey registry.Key
	opened := false
	for {
		if !opened {
			if k, err := registry.OpenKey(registry.LOCAL_MACHINE, serialCommKey, syscall.KEY_NOTIFY); err == nil {
				opened = true
				regKey = k
				notifyCh <- true
			}
			continue
		}
		if err := winapp.RegNotifyChangeKeyValue(regKey, 0, 0x00000001|0x00000004, 0, 0); err != nil {
			log.Panicln("RegNotifyChangeKeyValue:", err)
		}
		notifyCh <- true
	}
}

func AvailablePorts() ([]string, error) {

	root, err := registry.OpenKey(registry.LOCAL_MACHINE, serialCommKey, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}

	ks, err := root.ReadValueNames(0)
	if err != nil {
		return nil, err
	}
	var ports []string
	for _, k := range ks {
		port, _, err := root.GetStringValue(k)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func CheckPortAvailable(portName string) error {
	ports, err := AvailablePorts()
	if err != nil {
		return err
	}
	if len(ports) == 0 {
		return merry.New("нет доступных СОМ портов")
	}
	for _, s := range ports {
		if s == portName {
			return nil
		}
	}
	return merry.Errorf("%q: не допустимое имя ком порта: %v", portName, ports)
}

func FirstAvailablePortName() (string, error) {
	ports, err := AvailablePorts()
	if err != nil {
		return "", err
	}
	if len(ports) == 0 {
		return "", merry.New("нет доступных СОМ портов")
	}
	return ports[0], nil
}

func CheckPortName(portName string) error {
	if len(portName) == 0 {
		return merry.New("не задано имя СОМ порта")
	}
	availPorts, err := AvailablePorts()
	if err != nil {
		return merry.Append(err, "Не удалось получить список СОМ портов, представленных в системе")
	}

	if len(availPorts) == 0 {
		return merry.New("СОМ порты отсутствуют")
	}
	for _, s := range availPorts {
		if portName == s {
			return nil
		}
	}
	return merry.Errorf("СОМ порт %q не представлен в системе: %v", portName, availPorts)
}

const serialCommKey = `hardware\devicemap\serialcomm`
