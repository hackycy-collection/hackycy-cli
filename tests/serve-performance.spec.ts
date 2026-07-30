import type { ChildProcess } from 'node:child_process'
import { spawn } from 'node:child_process'
import { link as linkFile, mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises'
import { createServer } from 'node:net'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { expect, test } from '@playwright/test'
import { optimizeImage } from 'wasm-image-optimization'

let fixtureRoot: string
let server: ChildProcess
let serverUrl: URL

async function availablePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const candidate = createServer()
    candidate.once('error', reject)
    candidate.listen(0, '127.0.0.1', () => {
      const address = candidate.address()
      if (!address || typeof address === 'string') {
        reject(new Error('Could not allocate a test port'))
        return
      }
      candidate.close(error => error ? reject(error) : resolve(address.port))
    })
  })
}

async function waitForServer(url: URL): Promise<void> {
  const deadline = Date.now() + 15_000
  while (Date.now() < deadline) {
    try {
      const response = await fetch(new URL('/api/directory?path=', url))
      if (response.ok)
        return
    }
    catch {}
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  throw new Error('Serve performance fixture did not start')
}

async function createLinks(source: string, directory: string, count: number, name: (index: number) => string): Promise<void> {
  for (let start = 0; start < count; start += 100) {
    await Promise.all(Array.from(
      { length: Math.min(100, count - start) },
      (_, offset) => linkFile(source, path.join(directory, name(start + offset))),
    ))
  }
}

test.beforeAll(async () => {
  fixtureRoot = await mkdtemp(path.join(tmpdir(), 'ycy-serve-performance-'))
  const fixtureDirectory = path.join(fixtureRoot, '.fixtures')
  const imageDirectory = path.join(fixtureRoot, 'images')
  const itemDirectory = path.join(fixtureRoot, 'items')
  await Promise.all([
    mkdir(fixtureDirectory),
    mkdir(imageDirectory),
    mkdir(itemDirectory),
  ])

  const sourceSvg = new TextEncoder().encode('<svg xmlns="http://www.w3.org/2000/svg" width="3000" height="3000"><rect width="3000" height="3000" fill="#d9e4ea"/><circle cx="1500" cy="1500" r="900" fill="#137c8b"/></svg>')
  const image = await optimizeImage({ image: sourceSvg, width: 3000, height: 3000, format: 'jpeg', quality: 85 })
  const imageSource = path.join(fixtureDirectory, 'source.jpg')
  const itemSource = path.join(fixtureDirectory, 'source.txt')
  await Promise.all([
    writeFile(imageSource, image.data),
    writeFile(itemSource, 'performance fixture'),
  ])
  await Promise.all([
    createLinks(imageSource, imageDirectory, 78, index => `image-${String(index).padStart(3, '0')}.jpg`),
    createLinks(itemSource, itemDirectory, 1000, index => `item-${String(index).padStart(4, '0')}.txt`),
  ])

  const port = await availablePort()
  serverUrl = new URL(`http://127.0.0.1:${port}`)
  server = spawn('bun', ['src/cli.ts', 'serve', fixtureRoot, '--address', '127.0.0.1', '--port', String(port)], {
    cwd: path.resolve(import.meta.dirname, '..'),
    stdio: 'ignore',
  })
  await waitForServer(serverUrl)
})

test.afterAll(async () => {
  if (server && !server.killed)
    server.kill('SIGTERM')
  await rm(fixtureRoot, { recursive: true, force: true })
})

test.use({
  launchOptions: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH
    ? { executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH }
    : {},
  viewport: { width: 1280, height: 720 },
})

test('78 large images stay thumbnail-only and bounded while browsing', async ({ page }) => {
  const requests: string[] = []
  page.on('request', request => requests.push(request.url()))
  await page.addInitScript(() => {
    localStorage.setItem('ycy-serve-view', JSON.stringify('grid'))
    Object.assign(window, { __serveLongTasks: [] as number[] })
    new PerformanceObserver((list) => {
      const durations = (window as unknown as { __serveLongTasks: number[] }).__serveLongTasks
      durations.push(...list.getEntries().map(entry => entry.duration))
    }).observe({ type: 'longtask', buffered: true })
  })
  await page.route('**/thumbnails/images/image-000.jpg', route => route.fulfill({ status: 503, body: 'fixture failure' }))

  await page.goto(new URL('/browse/images', serverUrl).href)
  const fileArea = page.locator('.file-browser')
  await expect.poll(() => fileArea.locator('.grid-entry').count()).toBeGreaterThan(0)
  await expect.poll(() => fileArea.locator('img').evaluateAll(images => images.every((image) => {
    const thumbnail = image as HTMLImageElement
    return thumbnail.complete && thumbnail.naturalWidth > 0
  }))).toBe(true)

  const metrics = await fileArea.evaluate((root) => {
    const row = root.querySelector<HTMLElement>('.virtual-grid-row')
    const viewport = row?.closest<HTMLElement>('[data-radix-scroll-area-viewport]')
    const images = Array.from(root.querySelectorAll<HTMLImageElement>('img'))
    const columns = row ? getComputedStyle(row).gridTemplateColumns.split(' ').length : 0
    return {
      columns,
      decodedPixels: images.reduce((total, image) => total + image.naturalWidth * image.naturalHeight, 0),
      imageCount: images.length,
      maxImageHeight: Math.max(0, ...images.map(image => image.naturalHeight)),
      maxImageWidth: Math.max(0, ...images.map(image => image.naturalWidth)),
      mountedEntries: root.querySelectorAll('.grid-entry').length,
      mountedRows: root.querySelectorAll('.virtual-grid-row').length,
      viewportHeight: viewport?.clientHeight ?? 0,
    }
  })
  const maximumRows = Math.ceil(metrics.viewportHeight / 136) + 2
  expect(metrics.mountedRows).toBeLessThanOrEqual(maximumRows)
  expect(metrics.mountedEntries).toBeLessThanOrEqual(maximumRows * metrics.columns)
  expect(metrics.imageCount).toBeGreaterThan(0)
  expect(metrics.maxImageWidth).toBeLessThanOrEqual(160)
  expect(metrics.maxImageHeight).toBeLessThanOrEqual(160)
  expect(metrics.decodedPixels).toBeLessThan(1_000_000)
  expect(requests.filter(url => new URL(url).pathname.startsWith('/files/'))).toEqual([])

  const failedEntry = fileArea.getByRole('option').filter({ hasText: 'image-000.jpg' })
  await expect(failedEntry.locator('img')).toHaveCount(0)
  await expect(failedEntry.locator('svg')).toHaveCount(1)

  await page.evaluate(() => ((window as unknown as { __serveLongTasks: number[] }).__serveLongTasks = []))
  await page.getByRole('button', { name: 'Details view' }).click()
  await expect.poll(() => fileArea.locator('.details-row').count()).toBeGreaterThan(0)
  await page.getByRole('button', { name: 'Grid view' }).click()
  await expect.poll(() => fileArea.locator('.grid-entry').count()).toBeGreaterThan(0)
  await fileArea.locator('[data-radix-scroll-area-viewport]').evaluate(async (viewport) => {
    for (let offset = 0; offset <= viewport.scrollHeight; offset += 96) {
      viewport.scrollTop = offset
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()))
    }
  })
  const finalEntry = fileArea.getByRole('option').filter({ hasText: 'image-077.jpg' })
  await expect(finalEntry).toHaveCount(1)
  await finalEntry.click()
  const longTasks = await page.evaluate(() => (window as unknown as { __serveLongTasks: number[] }).__serveLongTasks)
  expect(Math.max(0, ...longTasks)).toBeLessThan(200)
})

test('1000-item main file area remains virtualized and keeps core interactions', async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('ycy-serve-view', JSON.stringify('list')))
  await page.goto(new URL('/browse/items', serverUrl).href)
  const fileArea = page.locator('.file-browser')
  const options = fileArea.getByRole('option')
  await expect.poll(() => options.count()).toBeGreaterThan(0)
  expect(await options.count()).toBeLessThanOrEqual(32)

  const viewport = fileArea.locator('[data-radix-scroll-area-viewport]')
  await viewport.evaluate(element => element.scrollTop = element.scrollHeight)
  const lastEntry = options.filter({ hasText: 'item-0999.txt' })
  await expect(lastEntry).toHaveCount(1)

  const search = page.getByRole('textbox', { name: 'Search current folder' })
  await search.fill('item-0500')
  await expect(options).toHaveCount(1)
  await expect(options.filter({ hasText: 'item-0500.txt' })).toHaveCount(1)
  await search.fill('')

  await page.getByRole('button', { name: 'Name' }).click()
  const descendingFirst = options.filter({ hasText: 'item-0999.txt' })
  await expect(descendingFirst).toHaveCount(1)
  await descendingFirst.click()
  const descendingSecond = options.filter({ hasText: 'item-0998.txt' })
  await expect(descendingSecond).toHaveCount(1)
  await descendingSecond.click({ modifiers: ['Shift'] })
  await expect(fileArea.locator('[role="option"][aria-selected="true"]')).toHaveCount(2)
  expect(await options.count()).toBeLessThanOrEqual(32)
})
