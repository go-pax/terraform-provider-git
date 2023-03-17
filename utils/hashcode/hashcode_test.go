package hashcode

import (
	"testing"
)

func TestString(t *testing.T) {
	v := "hello, world"
	expected := HashcodeString(v)
	for i := 0; i < 100; i++ {
		actual := HashcodeString(v)
		if actual != expected {
			t.Fatalf("bad: %#v\n\t%#v", actual, expected)
		}
	}
}

func TestString_positiveIndex(t *testing.T) {
	ips := []string{"192.168.1.3", "192.168.1.5"}
	for _, ip := range ips {
		if index := HashcodeString(ip); index < 0 {
			t.Fatalf("Bad Index %#v for ip %s", index, ip)
		}
	}
}
