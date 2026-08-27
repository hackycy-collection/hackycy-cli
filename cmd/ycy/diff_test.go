package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/netip"
	"reflect"
	"strings"
	"testing"

	diffcommand "github.com/hackycy/hackycy-cli/internal/commands/diff"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestTerminalDiffPresentationPreservesPlainAndAutomationReadiness(t *testing.T) {
	startup := diffcommand.Startup{
		LocalURL:          "http://localhost:43123",
		NetworkURLs:       []string{"http://192.168.1.50:43123", "http://10.0.0.8:43123"},
		BaselineDirectory: "/workspace/baseline",
		TargetDirectory:   "/workspace/target",
		Port:              43123,
	}
	want := "Directory diff: http://localhost:43123\n" +
		"MCP endpoint:   http://localhost:43123/mcp\n" +
		"Network: http://192.168.1.50:43123\n" +
		"Network MCP: http://192.168.1.50:43123/mcp\n" +
		"Network: http://10.0.0.8:43123\n" +
		"Network MCP: http://10.0.0.8:43123/mcp\n" +
		"Baseline: /workspace/baseline\n" +
		"Target:   /workspace/target\n"
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.PlainInteractive},
		{Kind: terminalexperience.Automation},
	} {
		var output, diagnostics bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output, Diagnostics: &diagnostics})
		run := experience.Open(context.Background())
		if err := run.Present(terminalDiffStartupDocument(session, startup)); err != nil {
			t.Fatalf("%v Present() error = %v", session.Kind, err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("%v Close() error = %v", session.Kind, err)
		}
		if output.String() != want {
			t.Fatalf("%v presentation = %q, want %q", session.Kind, output.String(), want)
		}
		if terminaltest.ContainsTerminalControl(output.Bytes()) {
			t.Fatalf("%v readiness contains terminal control: %q", session.Kind, output.String())
		}
		if diagnostics.Len() != 0 {
			t.Fatalf("%v readiness wrote stderr: %q", session.Kind, diagnostics.String())
		}
	}
}

func TestTerminalDiffPresentationUsesRichSemanticRoles(t *testing.T) {
	startup := diffcommand.Startup{
		LocalURL:          "http://localhost:43123",
		NetworkURLs:       []string{"http://192.168.1.50:43123"},
		BaselineDirectory: "/workspace/baseline",
		TargetDirectory:   "/workspace/target",
	}
	for _, session := range []terminalexperience.Session{
		{Kind: terminalexperience.RichInteractive, Color: true},
		{Kind: terminalexperience.RichInteractive},
	} {
		document := terminalDiffStartupDocument(session, startup)
		want := []terminalexperience.VisualRole{
			terminalexperience.VisualRoleActive,
			terminalexperience.VisualRoleMuted,
			terminalexperience.VisualRoleActive,
			terminalexperience.VisualRoleMuted,
			terminalexperience.VisualRoleMuted,
			terminalexperience.VisualRoleMuted,
		}
		if document.ClearBefore || len(document.Blocks) != len(want) {
			t.Fatalf("Rich document = %#v", document)
		}
		for index, role := range want {
			if document.Blocks[index].Role != role {
				t.Fatalf("Rich block %d role = %v, want %v", index, document.Blocks[index].Role, role)
			}
		}
		var output bytes.Buffer
		experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{Session: session, Output: &output})
		run := experience.Open(context.Background())
		if err := run.Present(document); err != nil {
			t.Fatalf("Present() error = %v", err)
		}
		if err := run.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if !session.Color && strings.Contains(output.String(), "\x1b[") {
			t.Fatalf("NO_COLOR Rich output contains style bytes: %q", output.String())
		}
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
