import process from 'node:process'
import * as p from '@clack/prompts'
import ansis from 'ansis'
import { printTitle } from '../../../shared/utils'
import { testAiProfile } from './client'
import {
  addAiProfile,
  listAiProfiles,
  removeAiProfile,
  resolveAiProfile,
  setAiProfileValue,
  setDefaultAiProfile,
} from './config'

export async function runAiConfigAdd(): Promise<void> {
  printTitle()
  p.intro(ansis.cyan('Git AI Config — Add Profile'))

  const result = await p.group(
    {
      name: () =>
        p.text({
          message: 'Profile name',
          placeholder: 'e.g. openai, deepseek, work',
          validate: (v = '') => {
            if (!v.trim())
              return 'Name is required'
            if (/\s/.test(v))
              return 'Name cannot contain spaces'
          },
        }),
      baseURL: () =>
        p.text({
          message: 'OpenAI-compatible base URL',
          placeholder: 'https://api.openai.com/v1',
          validate: (v = '') => {
            if (!v.trim())
              return 'Base URL is required'
          },
        }),
      model: () =>
        p.text({
          message: 'Model',
          placeholder: 'gpt-4.1-mini',
          validate: (v = '') => {
            if (!v.trim())
              return 'Model is required'
          },
        }),
      apiKey: () =>
        p.password({
          message: 'API key',
          validate: (v = '') => {
            if (!v.trim())
              return 'API key is required'
          },
        }),
    },
    {
      onCancel: () => {
        p.cancel('Cancelled')
        process.exit(0)
      },
    },
  )

  const spin = p.spinner()
  spin.start('Saving AI profile...')
  try {
    await addAiProfile(result.name, result.baseURL, result.model, result.apiKey)
    spin.stop('AI profile saved')
    p.outro(`Profile ${ansis.cyan(result.name)} added`)
  }
  catch (err) {
    spin.stop('Failed to save AI profile')
    p.log.error((err as Error).message)
    process.exit(1)
  }
}

export async function runAiConfigList(): Promise<void> {
  printTitle()
  console.log(ansis.dim('Git AI Config — Profiles'))
  console.log()

  const ai = await listAiProfiles()
  const entries = Object.entries(ai.profiles)
  if (entries.length === 0) {
    p.log.info('No AI profiles configured. Run "ycy git ai config add" to add one.')
    p.outro('')
    return
  }

  for (const [name, profile] of entries) {
    const marker = ai.defaultProfile === name ? ansis.green('*') : ' '
    console.log(`${marker} ${ansis.cyan(name)} ${ansis.dim(profile.model)} ${ansis.dim(profile.baseURL)}`)
  }
  console.log()
  p.outro('')
}

export async function runAiConfigUse(name: string): Promise<void> {
  try {
    await setDefaultAiProfile(name)
    p.outro(`Default AI profile set to ${ansis.cyan(name)}`)
  }
  catch (err) {
    p.log.error((err as Error).message)
    process.exit(1)
  }
}

export async function runAiConfigRemove(name: string): Promise<void> {
  const confirmed = await p.confirm({
    message: `Remove AI profile "${name}"?`,
  })

  if (p.isCancel(confirmed) || !confirmed) {
    p.outro('Cancelled')
    return
  }

  const removed = await removeAiProfile(name)
  if (!removed) {
    p.log.error(`AI profile not found: ${name}`)
    process.exit(1)
  }

  p.outro(`Profile ${ansis.cyan(name)} removed`)
}

export async function runAiConfigSet(name: string, key: string, value: string): Promise<void> {
  try {
    await setAiProfileValue(name, key, value)
    p.outro(`Profile ${ansis.cyan(name)} updated`)
  }
  catch (err) {
    p.log.error((err as Error).message)
    process.exit(1)
  }
}

export async function runAiConfigTest(name?: string): Promise<void> {
  printTitle()
  p.intro(ansis.cyan('Git AI Config — Test Profile'))

  let profile
  try {
    profile = await resolveAiProfile(name)
  }
  catch (err) {
    p.log.error((err as Error).message)
    process.exit(1)
  }

  const spin = p.spinner()
  spin.start(`Testing ${profile.name}...`)
  try {
    const content = await testAiProfile(profile)
    spin.stop('AI profile is reachable')
    p.log.info(`Response: ${content}`)
    p.outro('Done')
  }
  catch (err) {
    spin.stop('AI profile test failed')
    p.log.error((err as Error).message)
    p.log.info(`Provider: ${profile.name}`)
    p.log.info(`Base URL: ${profile.baseURL}`)
    p.log.info(`Model: ${profile.model}`)
    process.exit(1)
  }
}
