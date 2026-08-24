package main

import (
	"fmt"
	"io"
	"net"
	"net/netip"

	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
)

func newFSModule(output io.Writer) (*fscommand.Module, error) {
	return fscommand.New(fscommand.Dependencies{
		NetworkInterfaces: osFSNetworkInterfaces,
		Presenter:         terminalFSPresenter{output: output},
	})
}

func osFSNetworkInterfaces() ([]fscommand.NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	result := make([]fscommand.NetworkInterface, 0, len(interfaces))
	for _, network := range interfaces {
		addresses, err := network.Addrs()
		if err != nil {
			return nil, err
		}
		item := fscommand.NetworkInterface{Internal: network.Flags&net.FlagLoopback != 0}
		for _, address := range addresses {
			if parsed, ok := fsNetworkIP(address); ok {
				item.Addresses = append(item.Addresses, parsed)
			}
		}
		if len(item.Addresses) > 0 {
			result = append(result, item)
		}
	}
	return result, nil
}

func fsNetworkIP(address net.Addr) (netip.Addr, bool) {
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

type terminalFSPresenter struct {
	output io.Writer
}

func (presenter terminalFSPresenter) Present(start fscommand.Startup) error {
	if _, err := fmt.Fprintln(presenter.output, "File Browser"); err != nil {
		return err
	}
	for _, url := range start.URLs {
		if _, err := fmt.Fprintf(presenter.output, "%s: %s\n", url.Label, url.URL); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(presenter.output, "Directory: %s\nBind: %s:%d\nManagement: %t\nChunked uploads: %t\nHTML execution: %t\nAuthentication: %t\n", start.Directory, start.BindingAddress, start.Port, start.ManagementEnabled, start.ChunkedUploads, !start.SafeHTML, start.Authentication); err != nil {
		return err
	}
	if start.Authentication {
		_, err := fmt.Fprintf(presenter.output, "Session storage: %s\n", start.SessionDirectory)
		return err
	}
	return nil
}

func (presenter terminalFSPresenter) Stopped() error {
	_, err := fmt.Fprintln(presenter.output, "File Browser stopped.")
	return err
}
