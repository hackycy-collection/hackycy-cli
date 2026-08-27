package main

import (
	"context"
	"net"
	"net/netip"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/cliapp"
	diffcommand "github.com/hackycy/hackycy-cli/internal/commands/diff"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
)

func newDiffHandler(experience *terminalexperience.Runtime) (cliapp.DiffHandler, error) {
	module, err := diffcommand.New(diffcommand.Dependencies{
		NetworkInterfaces: osDiffNetworkInterfaces,
	})
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, input diffcommand.Input) (diffcommand.Result, error) {
		operation, err := module.Start(ctx, input)
		if err != nil || operation == nil {
			return diffcommand.Result{}, err
		}
		if ctx != nil && ctx.Err() != nil {
			return diffcommand.Result{}, operation.Wait(ctx)
		}
		run := experience.Open(ctx)
		defer run.Close()
		if err := run.Present(terminalDiffStartupDocument(experience.Session(), operation.Startup)); err != nil {
			_ = operation.Close()
			return diffcommand.Result{}, err
		}
		if err := operation.Wait(ctx); err != nil {
			return diffcommand.Result{}, err
		}
		return diffcommand.Result{}, nil
	}, nil
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

func terminalDiffStartupDocument(session terminalexperience.Session, start diffcommand.Startup) terminalexperience.PresentationDocument {
	if session.Kind != terminalexperience.RichInteractive {
		return terminalexperience.PresentationDocument{Blocks: []terminalexperience.PresentationBlock{{
			Role: terminalexperience.VisualRolePlain,
			Text: terminalDiffStartupPlainText(start),
		}}}
	}
	blocks := []terminalexperience.PresentationBlock{{
		Role: terminalexperience.VisualRoleActive,
		Text: "Directory diff: " + start.LocalURL,
	}, {
		Role: terminalexperience.VisualRoleMuted,
		Text: "MCP endpoint:   " + start.LocalURL + "/mcp",
	}}
	for _, url := range start.NetworkURLs {
		blocks = append(blocks,
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleActive, Text: "Network: " + url},
			terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Network MCP: " + url + "/mcp"},
		)
	}
	blocks = append(blocks,
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Baseline: " + start.BaselineDirectory},
		terminalexperience.PresentationBlock{Role: terminalexperience.VisualRoleMuted, Text: "Target:   " + start.TargetDirectory},
	)
	return terminalexperience.PresentationDocument{Blocks: blocks}
}

func terminalDiffStartupPlainText(start diffcommand.Startup) string {
	var output strings.Builder
	output.WriteString("Directory diff: ")
	output.WriteString(start.LocalURL)
	output.WriteString("\nMCP endpoint:   ")
	output.WriteString(start.LocalURL)
	output.WriteString("/mcp\n")
	for _, url := range start.NetworkURLs {
		output.WriteString("Network: ")
		output.WriteString(url)
		output.WriteString("\nNetwork MCP: ")
		output.WriteString(url)
		output.WriteString("/mcp\n")
	}
	output.WriteString("Baseline: ")
	output.WriteString(start.BaselineDirectory)
	output.WriteString("\nTarget:   ")
	output.WriteString(start.TargetDirectory)
	output.WriteByte('\n')
	return output.String()
}
