package main

import (
	"bytes"
	"errors"
	"net"
	"net/netip"
	"reflect"
	"testing"

	diffcommand "github.com/hackycy/hackycy-cli/internal/commands/diff"
)

func TestTerminalDiffPresenterWritesLegacyStartupURLs(t *testing.T) {
	output := &bytes.Buffer{}
	presenter := terminalDiffPresenter{output: output}
	err := presenter.Present(diffcommand.Startup{
		LocalURL:          "http://localhost:43123",
		NetworkURLs:       []string{"http://192.168.1.50:43123", "http://10.0.0.8:43123"},
		BaselineDirectory: "/workspace/baseline",
		TargetDirectory:   "/workspace/target",
		Port:              43123,
	})
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	want := "Directory diff: http://localhost:43123\n" +
		"MCP endpoint:   http://localhost:43123/mcp\n" +
		"Network: http://192.168.1.50:43123\n" +
		"Network MCP: http://192.168.1.50:43123/mcp\n" +
		"Network: http://10.0.0.8:43123\n" +
		"Network MCP: http://10.0.0.8:43123/mcp\n" +
		"Baseline: /workspace/baseline\n" +
		"Target:   /workspace/target\n"
	if output.String() != want {
		t.Fatalf("presentation = %q, want %q", output.String(), want)
	}
}

func TestCollectDiffNetworkInterfacesPreservesIPv4AndLoopbackFacts(t *testing.T) {
	interfaces := []net.Interface{
		{Index: 1, Name: "lo", Flags: net.FlagLoopback},
		{Index: 2, Name: "en0"},
		{Index: 3, Name: "broken"},
	}
	got, err := collectDiffNetworkInterfaces(interfaces, func(iface net.Interface) ([]net.Addr, error) {
		switch iface.Name {
		case "lo":
			return []net.Addr{&net.IPNet{IP: net.ParseIP("127.0.0.1")}}, nil
		case "en0":
			return []net.Addr{
				&net.IPNet{IP: net.ParseIP("192.168.1.50")},
				&net.IPAddr{IP: net.ParseIP("fe80::1")},
				fixtureNetAddr("not-an-ip"),
			}, nil
		default:
			return nil, errors.New("fixture failure")
		}
	})
	if err == nil || err.Error() != "fixture failure" || got != nil {
		t.Fatalf("error result = (%#v, %v)", got, err)
	}

	got, err = collectDiffNetworkInterfaces(interfaces[:2], func(iface net.Interface) ([]net.Addr, error) {
		if iface.Name == "lo" {
			return []net.Addr{&net.IPNet{IP: net.ParseIP("127.0.0.1")}}, nil
		}
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("192.168.1.50")},
			&net.IPAddr{IP: net.ParseIP("fe80::1")},
			fixtureNetAddr("not-an-ip"),
		}, nil
	})
	if err != nil {
		t.Fatalf("collectDiffNetworkInterfaces() error = %v", err)
	}
	want := []diffcommand.NetworkInterface{
		{Internal: true, Addresses: []netip.Addr{netip.MustParseAddr("127.0.0.1")}},
		{Addresses: []netip.Addr{netip.MustParseAddr("192.168.1.50"), netip.MustParseAddr("fe80::1")}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("interfaces = %#v, want %#v", got, want)
	}
}

type fixtureNetAddr string

func (address fixtureNetAddr) Network() string { return "fixture" }

func (address fixtureNetAddr) String() string { return string(address) }
