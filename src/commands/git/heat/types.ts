export type ChangeKind = 'M' | 'A' | 'D' | 'R' | 'C'
export type HeatTarget = 'files' | 'directories'

export interface GitHeatOptions {
  limit?: number
  days?: number
  type?: HeatTarget
}

export interface ChangeCounts {
  total: number
  modified: number
  added: number
  deleted: number
  renamed: number
  copied: number
}

export interface PathHeat extends ChangeCounts {
  path: string
}

export interface HeatReport {
  repoRoot: string
  repoName: string
  rangeLabel: string
  target: HeatTarget
  commitCount: number
  files: PathHeat[]
  directories: PathHeat[]
}
