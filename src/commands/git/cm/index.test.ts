import { describe, expect, test } from 'bun:test'
import { Command } from 'commander'
import { register } from './index'

describe('git cm command options', () => {
  test('keeps the supported command matrix', () => {
    const parent = new Command('git')
    register(parent)
    const command = parent.commands.find(item => item.name() === 'cm')

    expect(command?.options.map(option => option.long)).toEqual([
      '--profile',
      '--lang',
      '--staged',
      '--stage',
      '--stage-all',
      '--push',
      '--stage-push',
      '--dry-run',
      '--body',
    ])
  })
})
