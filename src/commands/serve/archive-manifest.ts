export const SEVEN_ZIP_VERSION = '26.02'
export const SEVEN_ZIP_RELEASE_BASE_URL = `https://github.com/ip7z/7zip/releases/download/${SEVEN_ZIP_VERSION}`

export type SevenZipTarget = 'darwin-arm64' | 'darwin-x64' | 'linux-arm64' | 'linux-x64' | 'win32-arm64' | 'win32-x64'

export interface SevenZipRuntimeFile {
  sourceName: string
  embeddedName: string
  filename: string
  sha256: string
  executable?: boolean
}

export interface SevenZipArtifact {
  asset: string
  sha256: string
  format: 'tar.xz' | 'windows-installer'
  files: SevenZipRuntimeFile[]
}

const UNIX_LICENSE: SevenZipRuntimeFile = {
  sourceName: 'License.txt',
  embeddedName: 'ycy-7zip-License.bin',
  filename: 'License.txt',
  sha256: '1790374e5352329cedb46ee3808930a88e9ca2f08b82b10fcf5cf605d2c301b1',
}

const WINDOWS_LICENSE: SevenZipRuntimeFile = {
  sourceName: 'License.txt',
  embeddedName: 'ycy-7zip-License.bin',
  filename: 'License.txt',
  sha256: '519ac0a4bded9c18ea02e0afb71f663d8c47373bd9facd3ac96a79f51d77765d',
}

export const SEVEN_ZIP_ARTIFACTS: Record<SevenZipTarget, SevenZipArtifact> = {
  'darwin-arm64': {
    asset: '7z2602-mac.tar.xz',
    sha256: '1cf6760579502f87e591ff5c73a005ec50b3e4d6f507e8b038382d563c3175b9',
    format: 'tar.xz',
    files: [
      { sourceName: '7zz', embeddedName: 'ycy-7zz.bin', filename: '7zz', sha256: '9c56cf3379a0d8544e9244958b96fdc7c17f9ce70f5a160eb2b41f5f3df96d8c', executable: true },
      UNIX_LICENSE,
    ],
  },
  'darwin-x64': {
    asset: '7z2602-mac.tar.xz',
    sha256: '1cf6760579502f87e591ff5c73a005ec50b3e4d6f507e8b038382d563c3175b9',
    format: 'tar.xz',
    files: [
      { sourceName: '7zz', embeddedName: 'ycy-7zz.bin', filename: '7zz', sha256: '9c56cf3379a0d8544e9244958b96fdc7c17f9ce70f5a160eb2b41f5f3df96d8c', executable: true },
      UNIX_LICENSE,
    ],
  },
  'linux-arm64': {
    asset: '7z2602-linux-arm64.tar.xz',
    sha256: '70ea6cc737ae1495ea2d7eb20ef3120fe579bd3f1a83a9d2362b62ec5bde2bba',
    format: 'tar.xz',
    files: [
      { sourceName: '7zz', embeddedName: 'ycy-7zz.bin', filename: '7zz', sha256: '41ca798f0c0652c435cbdd9c3ba49d703c9410c597f40a5cd336304b3964c674', executable: true },
      UNIX_LICENSE,
    ],
  },
  'linux-x64': {
    asset: '7z2602-linux-x64.tar.xz',
    sha256: '41aaba7b1235304ab5aa0624530c67ae829496cd29e875925271efdccc28c03e',
    format: 'tar.xz',
    files: [
      { sourceName: '7zz', embeddedName: 'ycy-7zz.bin', filename: '7zz', sha256: '1676a968815b92e865bc0ffeecee3fa284ba4402bf23dc2bec2412c4b502e922', executable: true },
      UNIX_LICENSE,
    ],
  },
  'win32-arm64': {
    asset: '7z2602-arm64.exe',
    sha256: '7c6fde79ed5e11b81c7bb6573b7962d3b6322aa5fce69c33ed19f672b55173ab',
    format: 'windows-installer',
    files: [
      { sourceName: '7z.exe', embeddedName: 'ycy-7z.exe', filename: '7z.exe', sha256: '46009c25732880c9d49032ec20da46dfdc669fb60257f50308a0026b4fac3aef', executable: true },
      { sourceName: '7z.dll', embeddedName: 'ycy-7z.dll', filename: '7z.dll', sha256: '7346eaea5f333b1d65b6b4eedf6797c416bbc91c75e46159df38aa28e153f7c5' },
      WINDOWS_LICENSE,
    ],
  },
  'win32-x64': {
    asset: '7z2602-x64.exe',
    sha256: '6745fa76dc2ea031596d8678f6f6b99c3c1b435b4164a63485adbbc7b8d82ef0',
    format: 'windows-installer',
    files: [
      { sourceName: '7z.exe', embeddedName: 'ycy-7z.exe', filename: '7z.exe', sha256: '83967f1b02b43c4efeda302795722c809e0e81b8307de73558d10484d5676a7d', executable: true },
      { sourceName: '7z.dll', embeddedName: 'ycy-7z.dll', filename: '7z.dll', sha256: '69fd4df057985c40e510e2fac182881c7f85e90aa13ec703f763a8fdb2ce61f8' },
      WINDOWS_LICENSE,
    ],
  },
}

export function sevenZipTarget(platform: NodeJS.Platform, architecture: string): SevenZipTarget | undefined {
  const candidate = `${platform}-${architecture}`
  return candidate in SEVEN_ZIP_ARTIFACTS ? candidate as SevenZipTarget : undefined
}
