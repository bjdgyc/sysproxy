package sysproxy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"sysproxy/shell"
)

func getInterfaceDisplayName(name string) (string, error) {
	content, err := shell.Exec("networksetup", "-listallhardwareports").ReadOutput()
	if err != nil {
		return "", err
	}
	for _, deviceSpan := range strings.Split(string(content), "Ethernet Address") {
		if strings.Contains(deviceSpan, "Device: "+name) {
			substr := "Hardware Port: "
			deviceSpan = deviceSpan[strings.Index(deviceSpan, substr)+len(substr):]
			deviceSpan = deviceSpan[:strings.Index(deviceSpan, "\n")]
			return deviceSpan, nil
		}
	}
	return "", errors.New(name + " not found in networksetup -listallhardwareports")
}

func SetSystemProxy(port uint16, isMixed bool, bypasses ...string) (clearFunc func() error, err error) {
	interfaceDisplayName, err := getNetworkInterface()
	if err != nil {
		return nil, err
	}
	if isMixed {
		err = shell.Exec("networksetup", "-setsocksfirewallproxy", interfaceDisplayName, "127.0.0.1", strconv.Itoa(int(port))).Attach().Run()
	}
	if err == nil {
		err = shell.Exec("networksetup", "-setwebproxy", interfaceDisplayName, "127.0.0.1", strconv.Itoa(int(port))).Attach().Run()
	}
	if err == nil {
		err = shell.Exec("networksetup", "-setsecurewebproxy", interfaceDisplayName, "127.0.0.1", strconv.Itoa(int(port))).Attach().Run()
	}

	return func() error {
		if isMixed {
			err = shell.Exec("networksetup", "-setsocksfirewallproxystate", interfaceDisplayName, "off").Attach().Run()
		}
		if err == nil {
			err = shell.Exec("networksetup", "-setwebproxystate", interfaceDisplayName, "off").Attach().Run()
		}
		if err == nil {
			err = shell.Exec("networksetup", "-setsecurewebproxystate", interfaceDisplayName, "off").Attach().Run()
		}
		return err
	}, err
}

func getNetworkInterface() (string, error) {
	buf, err := shell.Exec("sh", "-c", "networksetup -listnetworkserviceorder | grep -B 1 $(route -n get default | grep interface | awk '{print $2}')").ReadOutput()
	if err != nil {
		return "", err
	}
	reader := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(buf)))
	reg := regexp.MustCompile(`^\(\d+\)\s(.*)$`)
	device, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	match := reg.FindStringSubmatch(device)
	if len(match) <= 1 {
		return "", fmt.Errorf("unable to get network interface")
	}
	return match[1], nil
}
