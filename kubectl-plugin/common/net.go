package common

import (
	"errors"
	"fmt"
	"net"
)

// PortIsFree checks if the port is free or not
func PortIsFree(port int) bool {
	if listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port)); err == nil {
		_ = listener.Close()
		return true
	}

	return false
}

// FreePort is the tool to find free ports of KubeSphere
type FreePort struct {
	defaultPorts []int
}

// FindFreePortsOfKubeSphere returns the free ports of KubeSphere
// The order of the ports is: console, jenkins
func (f *FreePort) FindFreePortsOfKubeSphere() ([]int, error) {
	if len(f.defaultPorts) == 0 {
		f.defaultPorts = []int{30880, // for KubeSphere console
			30180, // for Jenkins
		}
	}

	conflict := false
	for _, port := range f.defaultPorts {
		if !PortIsFree(port) {
			conflict = true
			break
		}
	}

	if conflict {
		for i := range f.defaultPorts {
			f.defaultPorts[i]++
			if f.defaultPorts[i] >= 65530 {
				return nil, errors.New("no enough free ports available")
			}
		}
		return f.FindFreePortsOfKubeSphere()
	}
	return f.defaultPorts, nil
}
