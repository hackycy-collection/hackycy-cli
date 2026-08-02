import { describe, expect, test } from 'bun:test'
import { createServeAuthentication } from './authentication'

describe('ServeAuthentication', () => {
  test('parses account specifications and matches usernames without case sensitivity', async () => {
    const authentication = (await createServeAuthentication(['Alice:password:with-colon']))!

    const grant = await authentication.signIn({ username: 'ALICE', password: 'password:with-colon' })

    expect(grant?.account).toEqual({ username: 'Alice' })
    expect(authentication.resume(grant?.token)?.account).toEqual({ username: 'Alice' })
    authentication.close()
  })

  test('rejects invalid or duplicate account specifications before startup', async () => {
    await expect(createServeAuthentication(['alice-password123'])).rejects.toThrow('must use')
    await expect(createServeAuthentication(['bad name:password123'])).rejects.toThrow('Username must contain')
    await expect(createServeAuthentication(['alice:short'])).rejects.toThrow('Password must contain')
    await expect(createServeAuthentication(['Alice:password123', 'alice:password456'])).rejects.toThrow('specified more than once')
  })

  test('returns the same failure for an unknown username or wrong password', async () => {
    const authentication = (await createServeAuthentication(['alice:password123']))!

    expect(await authentication.signIn({ username: 'alice', password: 'incorrect-password' })).toBeUndefined()
    expect(await authentication.signIn({ username: 'missing', password: 'password123' })).toBeUndefined()
    authentication.close()
  })

  test('expires sessions and notifies active observers', async () => {
    const authentication = (await createServeAuthentication(['alice:password123'], { sessionLifetimeMs: 20 }))!
    const grant = (await authentication.signIn({ username: 'alice', password: 'password123' }))!
    let revoked = false
    authentication.observe(grant.token, () => {
      revoked = true
    })

    await Bun.sleep(40)

    expect(authentication.resume(grant.token)).toBeUndefined()
    expect(revoked).toBe(true)
    authentication.close()
  })

  test('evicts least-recently-used sessions at account and process limits', async () => {
    const authentication = (await createServeAuthentication(
      ['alice:password123', 'bob:password456'],
      { maxAccountSessions: 2, maxSessions: 3 },
    ))!
    const aliceFirst = (await authentication.signIn({ username: 'alice', password: 'password123' }))!
    const aliceSecond = (await authentication.signIn({ username: 'alice', password: 'password123' }))!
    authentication.resume(aliceFirst.token)
    const aliceThird = (await authentication.signIn({ username: 'alice', password: 'password123' }))!

    expect(authentication.resume(aliceSecond.token)).toBeUndefined()
    expect(authentication.resume(aliceFirst.token)).toBeDefined()
    expect(authentication.resume(aliceThird.token)).toBeDefined()

    const bobFirst = (await authentication.signIn({ username: 'bob', password: 'password456' }))!
    const bobSecond = (await authentication.signIn({ username: 'bob', password: 'password456' }))!

    expect(authentication.resume(aliceFirst.token)).toBeUndefined()
    expect(authentication.resume(aliceThird.token)).toBeDefined()
    expect(authentication.resume(bobFirst.token)).toBeDefined()
    expect(authentication.resume(bobSecond.token)).toBeDefined()
    authentication.close()
  })
})
