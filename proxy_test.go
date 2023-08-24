package sysproxy

import "testing"

func TestSetSystemProxy(t *testing.T) {
	clearFunc, err := SetSystemProxy(8088, true)
	if err != nil {
		t.Fatal(err)
	}
	err = clearFunc()
	if err != nil {
		t.Error(err)
	}
}

func TestSetSystemProxyWithBypass(t *testing.T) {
	_, err := SetSystemProxy(8088, true, "127.0.0.1", "192.168.0.1")
	if err != nil {
		t.Fatal(err)
	}
}
