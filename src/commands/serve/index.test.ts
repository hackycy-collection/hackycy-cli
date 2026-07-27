import { describe, expect, test } from 'bun:test'
import { Command } from 'commander'
import { register } from './index'

describe('serve command registration', () => {
  test('makes the directory optional and documents its default', () => {
    const program = new Command()
    register(program)

    const command = program.commands[0]!
    expect(command.usage()).toBe('[options] [directory]')
    expect(command.description()).toContain('defaults to current directory')
  })
})
