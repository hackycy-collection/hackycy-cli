import type { ResolvedAiProfile } from './types'

interface ChatMessage {
  role: 'system' | 'user'
  content: string
}

interface ChatCompletionResponse {
  choices?: Array<{
    message?: {
      content?: string
    }
  }>
}

export async function createChatCompletion(
  profile: ResolvedAiProfile,
  messages: ChatMessage[],
): Promise<string> {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), profile.timeoutMs)

  try {
    const res = await fetch(`${profile.baseURL}/chat/completions`, {
      method: 'POST',
      signal: controller.signal,
      headers: {
        'Authorization': `Bearer ${profile.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        model: profile.model,
        temperature: profile.temperature,
        max_tokens: profile.maxOutputTokens,
        messages,
      }),
    })

    if (!res.ok) {
      const text = await res.text()
      throw new Error(`${res.status} ${res.statusText}${text ? `: ${text}` : ''}`)
    }

    const json = await res.json() as ChatCompletionResponse
    const content = json.choices?.[0]?.message?.content?.trim()
    if (!content)
      throw new Error('AI returned an empty response')

    return content
  }
  catch (err) {
    if ((err as Error).name === 'AbortError')
      throw new Error(`AI request timed out after ${profile.timeoutMs}ms`)
    throw err
  }
  finally {
    clearTimeout(timeout)
  }
}

export async function testAiProfile(profile: ResolvedAiProfile): Promise<string> {
  const content = await createChatCompletion(profile, [
    {
      role: 'system',
      content: 'Return exactly: ok',
    },
    {
      role: 'user',
      content: 'Connection test.',
    },
  ])
  return content
}
