package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	rootcommand "github.com/hackycy/hackycy-cli/pkg/cmd/root"
)

func newFSHandler(experience *terminalexperience.Runtime) (rootcommand.FSHandler, error) {
	module, err := fscommand.New(fscommand.Dependencies{
		NetworkInterfaces: osFSNetworkInterfaces,
	})
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, input fscommand.Input) (fscommand.Result, error) {
		operation, err := module.Start(ctx, input)
		if err != nil || operation == nil {
			return fscommand.Result{}, err
		}
		if ctx != nil && ctx.Err() != nil {
			return fscommand.Result{}, operation.Wait(ctx)
		}
		run := experience.Open(ctx)
		defer run.Close()
		if err := run.Present(terminalFSStartupDocument(experience.Session(), operation.Startup)); err != nil {
			_ = operation.Close()
			return fscommand.Result{}, err
		}
		if err := operation.Wait(ctx); err != nil {
			return fscommand.Result{}, err
		}
		if err := run.Present(terminalFSStoppedDocument(experience.Session())); err != nil {
			return fscommand.Result{}, err
		}
		return fscommand.Result{}, nil
	}, nil
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

func terminalFSStartupDocument(session terminalexperience.Session, start fscommand.Startup) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: terminalFSStartupPlainText(start),
		}}}
	}
	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleActive,
		Text: "File Browser",
	}}
	for _, url := range start.URLs {
		blocks = append(blocks, terminalexperience.PresentationBlock{
			Role: terminalexperience.VisualRoleActive,
			Text: url.Label + ": " + url.URL,
		})
	}
	blocks = append(blocks,
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Directory: " + start.Directory},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Bind: %s:%d", start.BindingAddress, start.Port)},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Management: %t", start.ManagementEnabled)},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Chunked uploads: %t", start.ChunkedUploads)},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("HTML execution: %t", !start.SafeHTML)},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: fmt.Sprintf("Authentication: %t", start.Authentication)},
	)
	if start.Authentication {
		blocks = append(blocks, terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Session storage: " + start.SessionDirectory})
	}
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalFSStoppedDocument(session terminalexperience.Session) terminalexperience.PresentationDocument {
	role := terminalexperience.VisualRolePlain
	if session.Kind == terminalexperience.RichInteractive {
		role = terminalexperience.VisualRoleSuccess
	}
	return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
		Role: role,
		Text: "File Browser stopped.",
	}}}
}

func terminalFSStartupPlainText(start fscommand.Startup) string {
	var output strings.Builder
	output.WriteString("File Browser\n")
	for _, url := range start.URLs {
		_, _ = fmt.Fprintf(&output, "%s: %s\n", url.Label, url.URL)
	}
	_, _ = fmt.Fprintf(&output, "Directory: %s\nBind: %s:%d\nManagement: %t\nChunked uploads: %t\nHTML execution: %t\nAuthentication: %t\n", start.Directory, start.BindingAddress, start.Port, start.ManagementEnabled, start.ChunkedUploads, !start.SafeHTML, start.Authentication)
	if start.Authentication {
		_, _ = fmt.Fprintf(&output, "Session storage: %s\n", start.SessionDirectory)
	}
	return output.String()
}
