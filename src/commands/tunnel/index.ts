import type { Command } from 'commander'
import type { ClientOptionInput, ServerOptionInput } from './config'
import { resolveClientConfig, resolveServerConfig } from './config'

export function register(program: Command): void {
  const tunnel = program.command('tunnel').description('Manage trusted tunnel clients and tunnel definitions')

  tunnel
    .command('server')
    .description('Run the Tunnel Control Plane and supervised frps process')
    .option('--address <address>', 'Address for all tunnel server listeners')
    .option('--control-port <port>', 'Tunnel Control Plane port')
    .option('--frp-port <port>', 'FRP bind port')
    .option('--http-port <port>', 'FRP HTTP vhost port')
    .option('--port-range <start-end>', 'Server Port Pool')
    .option('--advertise-frp-addr <host:port>', 'FRP endpoint advertised to trusted clients')
    .option('--data-dir <path>', 'Tunnel server state directory')
    .action(async (options: ServerOptionInput) => {
      const { runTunnelServer } = await import('./server/run')
      await runTunnelServer(resolveServerConfig(options))
    })

  tunnel
    .command('connect')
    .description('Connect a native trusted client to a Tunnel Control Plane')
    .option('--server <control-plane>', 'Tunnel Control Plane origin')
    .option('--token <client-token>', 'Client Token')
    .action(async (options: ClientOptionInput) => {
      const { runTunnelClient } = await import('./client/run')
      await runTunnelClient(await resolveClientConfig(options))
    })
}
