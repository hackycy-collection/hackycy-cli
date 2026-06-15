import type { AiConfig, AiProfile, ResolvedAiProfile } from './types'
import process from 'node:process'
import { readConfig, writeConfig } from '../fork/config'
import { decrypt, deriveKey, encrypt } from '../fork/crypto'

const DEFAULT_TEMPERATURE = 0.2
const DEFAULT_TIMEOUT_MS = 30_000
const DEFAULT_MAX_OUTPUT_TOKENS = 300

const ENV_PROFILE = 'YCY_AI_PROFILE'
const ENV_BASE_URL = 'YCY_AI_BASE_URL'
const ENV_API_KEY = 'YCY_AI_API_KEY'
const ENV_MODEL = 'YCY_AI_MODEL'
const ENV_TEMPERATURE = 'YCY_AI_TEMPERATURE'
const ENV_TIMEOUT_MS = 'YCY_AI_TIMEOUT_MS'
const ENV_MAX_OUTPUT_TOKENS = 'YCY_AI_MAX_OUTPUT_TOKENS'

function emptyAiConfig(): AiConfig {
  return { profiles: {} }
}

function normalizeBaseURL(baseURL: string): string {
  return baseURL.trim().replace(/\/+$/, '')
}

function parseNumberEnv(value: string | undefined): number | undefined {
  if (!value)
    return undefined
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function withDefaults(profile: AiProfile): Omit<ResolvedAiProfile, 'name' | 'apiKey'> {
  return {
    baseURL: normalizeBaseURL(profile.baseURL),
    model: profile.model,
    temperature: profile.temperature ?? DEFAULT_TEMPERATURE,
    timeoutMs: profile.timeoutMs ?? DEFAULT_TIMEOUT_MS,
    maxOutputTokens: profile.maxOutputTokens ?? DEFAULT_MAX_OUTPUT_TOKENS,
  }
}

export async function readAiConfig(): Promise<AiConfig> {
  const config = await readConfig()
  if (!config.ai)
    return emptyAiConfig()
  return {
    ...config.ai,
    profiles: config.ai.profiles ?? {},
  }
}

export async function writeAiConfig(ai: AiConfig): Promise<void> {
  const config = await readConfig()
  config.ai = ai
  await writeConfig(config)
}

export async function addAiProfile(
  name: string,
  baseURL: string,
  model: string,
  apiKey: string,
): Promise<void> {
  const config = await readConfig()
  const ai = config.ai ?? emptyAiConfig()
  const key = await deriveKey(config.salt)
  ai.profiles[name] = {
    baseURL: normalizeBaseURL(baseURL),
    model: model.trim(),
    apiKey: encrypt(apiKey, key),
  }
  ai.defaultProfile ??= name
  config.ai = ai
  await writeConfig(config)
}

export async function removeAiProfile(name: string): Promise<boolean> {
  const ai = await readAiConfig()
  if (!ai.profiles[name])
    return false

  delete ai.profiles[name]
  if (ai.defaultProfile === name)
    ai.defaultProfile = Object.keys(ai.profiles)[0]

  await writeAiConfig(ai)
  return true
}

export async function setDefaultAiProfile(name: string): Promise<void> {
  const ai = await readAiConfig()
  if (!ai.profiles[name])
    throw new Error(`AI profile not found: ${name}`)
  ai.defaultProfile = name
  await writeAiConfig(ai)
}

export async function setAiProfileValue(name: string, key: string, value: string): Promise<void> {
  const config = await readConfig()
  const ai = config.ai ?? emptyAiConfig()
  const profile = ai.profiles[name]
  if (!profile)
    throw new Error(`AI profile not found: ${name}`)

  if (key === 'baseURL') {
    profile.baseURL = normalizeBaseURL(value)
  }
  else if (key === 'model') {
    profile.model = value.trim()
  }
  else if (key === 'apiKey') {
    const cryptoKey = await deriveKey(config.salt)
    profile.apiKey = encrypt(value, cryptoKey)
  }
  else if (key === 'temperature') {
    const parsed = Number(value)
    if (!Number.isFinite(parsed) || parsed < 0 || parsed > 2)
      throw new Error('temperature must be a number between 0 and 2')
    profile.temperature = parsed
  }
  else if (key === 'timeoutMs') {
    const parsed = Number.parseInt(value, 10)
    if (!Number.isInteger(parsed) || parsed < 1000)
      throw new Error('timeoutMs must be an integer greater than or equal to 1000')
    profile.timeoutMs = parsed
  }
  else if (key === 'maxOutputTokens') {
    const parsed = Number.parseInt(value, 10)
    if (!Number.isInteger(parsed) || parsed < 32)
      throw new Error('maxOutputTokens must be an integer greater than or equal to 32')
    profile.maxOutputTokens = parsed
  }
  else {
    throw new Error('Unsupported key. Use baseURL, model, apiKey, temperature, timeoutMs, or maxOutputTokens.')
  }

  config.ai = ai
  await writeConfig(config)
}

export async function listAiProfiles(): Promise<AiConfig> {
  return readAiConfig()
}

export async function resolveAiProfile(profileName?: string): Promise<ResolvedAiProfile> {
  const config = await readConfig()
  const ai = config.ai ?? emptyAiConfig()
  const env = process.env
  const selectedName = profileName || env[ENV_PROFILE] || ai.defaultProfile || Object.keys(ai.profiles)[0]
  const stored = selectedName ? ai.profiles[selectedName] : undefined

  const baseURL = env[ENV_BASE_URL] || stored?.baseURL
  const model = env[ENV_MODEL] || stored?.model
  let apiKey = env[ENV_API_KEY]

  if (!apiKey && stored?.apiKey) {
    const key = await deriveKey(config.salt)
    apiKey = decrypt(stored.apiKey, key)
  }

  if (!baseURL || !model || !apiKey) {
    throw new Error('No usable AI profile found. Run "ycy git ai config add" or set YCY_AI_BASE_URL, YCY_AI_MODEL, and YCY_AI_API_KEY.')
  }

  const defaults = stored
    ? withDefaults(stored)
    : {
        baseURL: normalizeBaseURL(baseURL),
        model,
        temperature: DEFAULT_TEMPERATURE,
        timeoutMs: DEFAULT_TIMEOUT_MS,
        maxOutputTokens: DEFAULT_MAX_OUTPUT_TOKENS,
      }

  return {
    name: selectedName || 'env',
    baseURL: normalizeBaseURL(baseURL),
    model,
    apiKey,
    temperature: parseNumberEnv(env[ENV_TEMPERATURE]) ?? defaults.temperature,
    timeoutMs: parseNumberEnv(env[ENV_TIMEOUT_MS]) ?? defaults.timeoutMs,
    maxOutputTokens: parseNumberEnv(env[ENV_MAX_OUTPUT_TOKENS]) ?? defaults.maxOutputTokens,
  }
}
