package main

import (
	"fmt"
	"io"
	"net"
	"net/netip"

	diffcommand "github.com/hackycy/hackycy-cli/internal/commands/diff"
)

func newDiffModule(output io.Writer) (*diffcommand.Module, error) {
	return diffcommand.New(diffcommand.Dependencies{
		NetworkInterfaces: osDiffNetworkInterfaces,
		Presenter:         terminalDiffPresenter{output: output},
	})
}

func osDiffNetworkInterfaces() ([]diffcommand.NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	return collectDiffNetworkInterfaces(interfaces, func(network net.Interface) ([]net.Addr, error) {
		return network.Addrs()
	})
}

func collectDiffNetworkInterfaces(interfaces []net.Interface, addresses func(net.Interface) ([]net.Addr, error)) ([]diffcommand.NetworkInterface, error) {
	result := make([]diffcommand.NetworkInterface, 0, len(interfaces))
	for _, network := range interfaces {
		observed, err := addresses(network)
		if err != nil {
			return nil, err
		}
		item := diffcommand.NetworkInterface{Internal: network.Flags&net.FlagLoopback != 0}
		for _, address := range observed {
			if ip, ok := diffNetworkIP(address); ok {
				item.Addresses = append(item.Addresses, ip)
			}
		}
		if len(item.Addresses) > 0 {
			result = append(result, item)
		}
	}
	return result, nil
}

func diffNetworkIP(address net.Addr) (netip.Addr, bool) {
	var raw net.IP
	switch value := address.(type) {
	case *net.IPNet:
		raw = value.IP
	case *net.IPAddr:
		raw = value.IP
	default:
		return netip.Addr{}, false
	}
	parsed, ok := netip.AddrFromSlice(raw)
	if !ok {
		return netip.Addr{}, false
	}
	return parsed.Unmap(), true
}

type terminalDiffPresenter struct {
	output io.Writer
}

func (presenter terminalDiffPresenter) Present(start diffcommand.Startup) error {
	if _, err := fmt.Fprintf(presenter.output, "Directory diff: %s\n", start.LocalURL); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(presenter.output, "MCP endpoint:   %s/mcp\n", start.LocalURL); err != nil {
		return err
	}
	for _, url := range start.NetworkURLs {
		if _, err := fmt.Fprintf(presenter.output, "Network: %s\nNetwork MCP: %s/mcp\n", url, url); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(presenter.output, "Baseline: %s\nTarget:   %s\n", start.BaselineDirectory, start.TargetDirectory); err != nil {
		return err
	}
	return nil
}
