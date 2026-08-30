# G9 Native Windows Host Availability

- Commit requiring native execution: `e916cb1ad5b7b5ef5a5b949647e0145208adb167`.
- Current controller host: Darwin arm64.
- Local VM command checks: no `VBoxManage`, `multipass`, `qemu-system-x86_64`, `qemu-system-aarch64`, `vmrun`, `prlctl`, or `limactl` executable was available.
- Local VM application checks: no UTM, VirtualBox, VMware Fusion, or Parallels application was present at its standard `/Applications` path.
- Docker server: `linux/arm64 28.1.1`; it supplies the completed native Linux evidence but cannot run Windows containers on this host.
- SSH configuration: no host entry explicitly named `windows` or `win` was present.

Result: UNAVAILABLE. No native Windows execution environment is available in the current task context. Cross-built PE artifacts remain compile evidence only and do not satisfy G9 E1 or E2.

Recovery input: a native Windows x64 or arm64 host with Go 1.26.7 and the exact commit checked out, able to run both `GOTOOLCHAIN=go1.26.7 GOWORK=off CGO_ENABLED=0 go test -count=1 ./...` and `make acceptance`. Its records must include commit, Windows version, architecture, Go version, command, result, and postflight source status under this evidence directory. Do not modify source; if source changes, rerun G8 and restart the full G9 set.
