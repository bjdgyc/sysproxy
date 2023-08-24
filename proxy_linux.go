package sysproxy

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sysproxy/shell"
)

var (
	hasGSettings bool
	sudoUser     string
)

func init() {
	_, err := exec.LookPath("gsettings")
	hasGSettings = err == nil
	if os.Getuid() == 0 {
		sudoUser = os.Getenv("SUDO_USER")
	}
}

func runAsUser(name string, args ...string) error {
	if os.Getuid() != 0 {
		return shell.Exec(name, args...).Attach().Run()
	} else if sudoUser != "" {
		return shell.Exec("su", "-", sudoUser, "-c", fmt.Sprint(name, " ", strings.Join(args, " "))).Attach().Run()
	} else {
		return errors.New("set system proxy: unable to set as root")
	}
}

// SetSystemProxy 设置系统代理
// isMixed: 是否支持多种代理
func SetSystemProxy(port uint16, isMixed bool, bypasses ...string) (clearFunc func() error, err error) {
	if !hasGSettings {
		return nil, errors.New("unsupported desktop environment")
	}
	err = runAsUser("gsettings", "set", "org.gnome.system.proxy.http", "enabled", "true")
	if err != nil {
		return nil, err
	}

	if isMixed {
		err = setGnomeProxy(port, "ftp", "http", "https", "socks")
	} else {
		err = setGnomeProxy(port, "http", "https")
	}
	if err != nil {
		return nil, err
	}
	err = runAsUser("gsettings", "set", "org.gnome.system.proxy", "use-same-proxy", strconv.FormatBool(isMixed))
	if err != nil {
		return nil, err
	}
	err = runAsUser("gsettings", "set", "org.gnome.system.proxy", "mode", "manual")
	if err != nil {
		return nil, err
	}
	return func() error {
		return runAsUser("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	}, nil
}

func setGnomeProxy(port uint16, proxyTypes ...string) error {
	for _, proxyType := range proxyTypes {
		err := runAsUser("gsettings", "set", "org.gnome.system.proxy."+proxyType, "host", "127.0.0.1")
		if err != nil {
			return err
		}
		err = runAsUser("gsettings", "set", "org.gnome.system.proxy."+proxyType, "port", strconv.Itoa(int(port)))
		if err != nil {
			return err
		}
	}
	return nil
}
