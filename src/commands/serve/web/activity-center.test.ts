import type { DownloadTask } from './api'
import { describe, expect, test } from 'bun:test'
import { downloadDetail, formatBytes } from './components/activity-center'

function task(update: Partial<DownloadTask>): DownloadTask {
  return {
    id: 'task',
    url: 'https://example.test/file',
    directoryPath: '',
    status: 'running',
    bytesDownloaded: 0,
    createdAt: '2026-01-01T00:00:00.000Z',
    ...update,
  }
}

describe('download activity presentation', () => {
  test('shows transferred bytes, total size, and speed for active downloads', () => {
    expect(downloadDetail(task({
      bytesDownloaded: 5 * 1024 * 1024,
      totalBytes: 20 * 1024 * 1024,
      speedBytesPerSecond: 2 * 1024 * 1024,
    }))).toBe('5.0 MiB / 20 MiB · 2.0 MiB/s')
  })

  test('keeps unknown totals and terminal states readable', () => {
    expect(downloadDetail(task({ bytesDownloaded: 1536 }))).toBe('1.5 KiB')
    expect(downloadDetail(task({ status: 'cancelled' }))).toBe('Cancelled')
    expect(downloadDetail(task({ status: 'error', error: 'Remote server returned HTTP 503' }))).toBe('Remote server returned HTTP 503')
    expect(formatBytes(0)).toBe('0 B')
  })
})
