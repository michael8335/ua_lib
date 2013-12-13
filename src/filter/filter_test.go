package filter

import (
	"fmt"
	"testing"
)

func TestSnsfilters(t *testing.T) {
	connstring := "ptmind:ptmind2012@tcp(192.168.16.51:3308)/ptmind_common?charset=utf8"
	filter := NewSnsfilter()
	filter.ReadInfo(connstring)

	tmp := map[string]string{"ref": "http://chiebukuro.yahoo.co.jp/test?test=go"}
	filter.Filt(tmp)
	fmt.Printf("TestSnsfilters: %v\n", tmp)
	if value, ok := tmp["rt"]; !ok || value != "chiebukuro.yahoo.co." {
		t.Error("Sns filter failed")
	}
}

func TestOsfilters(t *testing.T) {
	connstring := "ptmind:ptmind2012@tcp(192.168.16.51:3308)/ptmind_common?charset=utf8"
	filter := NewOsfilter()
	filter.ReadInfo(connstring)

	tmp := map[string]string{"ua": "Mozilla/5.0 (Linux; U; Android 4.1.1; zh-CN; G_3E_5 Build/JRO03C) AppleWebKit/534.31 (KHTML, like Gecko) UCBrowser/8.8.3.278 U3/0.8.0 Mobile Safari/534.31"}
	filter.Filt(tmp)
	fmt.Printf("TestOsfilters: %v\n", tmp)
}

func TestBrowserfilters(t *testing.T) {
	connstring := "ptmind:ptmind2012@tcp(192.168.16.51:3308)/ptmind_common?charset=utf8"
	filter := NewBrowserfilter()
	filter.ReadInfo(connstring)

	tmp := map[string]string{"ua": "Mozilla/5.0 (Linux; U; Android 4.1.1; zh-CN; G_3E_5 Build/JRO03C) AppleWebKit/534.31 (KHTML, like Gecko) UCBrowser/8.8.3.278 U3/0.8.0 Mobile Safari/534.31"}
	filter.Filt(tmp)
	fmt.Printf("TestBrowserfilters: %v\n", tmp)
}

func TestUAfilters(t *testing.T) {
	filter := NewUafilter()

	tmp := map[string]string{"ua": "Mozilla/5.0 (Linux; U; Android 4.1.1; zh-CN; G_3E_5 Build/JRO03C) AppleWebKit/534.31 (KHTML, like Gecko) UCBrowser/8.8.3.278 U3/0.8.0 Mobile Safari/534.31"}
	filter.Filt(tmp)
	fmt.Printf("TestUAfilters: %v\n", tmp)
}

func TestInstationfilters(t *testing.T) {
	filter := NewInstationfilter()

	tmp := map[string]string{"ref_type": "0"}
	filter.Filt(tmp)
	fmt.Printf("TestInstationfilters: %v\n", tmp)
}

func TestOutstationfilters(t *testing.T) {
	filter := NewOutstationfilter()

	tmp := map[string]string{"ref_type": "1", "ref": "http://chiebukuro.yahoo.co.jp/test?test=go"}
	filter.Filt(tmp)
	fmt.Printf("TestOutstationfilters: %v\n", tmp)
}
