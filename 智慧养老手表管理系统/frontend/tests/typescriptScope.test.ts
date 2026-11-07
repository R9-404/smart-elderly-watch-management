// @vitest-environment node
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'
import { afterEach, expect, it } from 'vitest'

const roots: string[] = []
const gate = fileURLToPath(new URL('../scripts/check-typescript-scope.mjs', import.meta.url))
function fixture() {
  const root = mkdtempSync(path.join(tmpdir(), 'suining-ts-scope-'))
  roots.push(root)
  mkdirSync(path.join(root, 'src'))
  const config = { compilerOptions: { strict: true, allowJs: false }, include: ['src/**/*.ts', 'src/**/*.vue'] }
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify(config))
  writeFileSync(path.join(root, 'src', 'App.vue'), '<template><div /></template><script setup lang="ts">const value: number = 1</script>')
  return { root, config }
}
function check(root: string) {
  return spawnSync(process.execPath, [gate, root], { encoding: 'utf8', windowsHide: true, timeout: 10000 })
}
afterEach(() => {
  for (const root of roots.splice(0)) {
    if (path.dirname(root) !== path.resolve(tmpdir()) || !path.basename(root).startsWith('suining-ts-scope-')) {
      throw new Error('Unexpected fixture cleanup path')
    }
    rmSync(root, { recursive: true, force: true })
  }
})
it('accepts a complete typed component scope', () => {
  const { root } = fixture()
  expect(check(root).status).toBe(0)
})
it('rejects a Vue script that silently escapes type checking', () => {
  const { root } = fixture()
  writeFileSync(path.join(root, 'src', 'App.vue'), '<template><div /></template><script setup>const untyped = []</script>')
  const result = check(root)
  expect(result.status).toBe(1)
  expect(result.stderr).toContain('inline lang="ts"')
})
it('rejects a JavaScript module reintroduced under src', () => {
  const { root } = fixture()
  writeFileSync(path.join(root, 'src', 'hidden.js'), 'export default 1')
  expect(check(root).status).toBe(1)
})
it('rejects narrowing the include list to a whitelist', () => {
  const { root, config } = fixture()
  config.include = ['src/App.vue']
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify(config))
  expect(check(root).status).toBe(1)
})
it('rejects disabling implicit-any checking under strict:true', () => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify({ ...config, compilerOptions: { ...config.compilerOptions, noImplicitAny: false } }))
  expect(check(root).status).toBe(1)
})
it('rejects excluded source files', () => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify({ ...config, exclude: ['src/App.vue'] }))
  expect(check(root).status).toBe(1)
})
it('rejects a typed file extension outside the actual compiler include set', () => {
  const { root } = fixture()
  writeFileSync(path.join(root, 'src', 'hidden.tsx'), 'export const value: number = 1')
  const result = check(root)
  expect(result.status).toBe(1)
  expect(result.stderr).toContain('absent from the compiler include set')
})

it('accepts inherited strict options while still covering every component', () => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, 'base.json'), JSON.stringify({ compilerOptions: { strict: true, allowJs: false } }))
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify({ extends: './base.json', include: config.include }))
  expect(check(root).status).toBe(0)
})

it('rejects strict checking disabled in an inherited configuration', () => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, 'base.json'), JSON.stringify({ compilerOptions: { strict: false, allowJs: false } }))
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify({ extends: './base.json', include: config.include }))
  expect(check(root).status).toBe(1)
})

it('rejects noCheck hidden in an inherited configuration', () => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, 'base.json'), JSON.stringify({ compilerOptions: { ...config.compilerOptions, noCheck: true } }))
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify({ extends: './base.json', include: config.include }))
  expect(check(root).status).toBe(1)
})

it.each([
  { baseUrl: '.' },
  { ignoreDeprecations: '6.0' }
])('rejects deprecated configuration or a migration warning bypass: %j', options => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, 'tsconfig.app.json'), JSON.stringify({
    ...config, compilerOptions: { ...config.compilerOptions, ...options }
  }))
  expect(check(root).status).toBe(1)
})

it.each(['tsconfig.node.json', 'tsconfig.test.json'])('rejects type checking disabled in %s', name => {
  const { root, config } = fixture()
  writeFileSync(path.join(root, name), JSON.stringify({
    ...config, compilerOptions: { ...config.compilerOptions, noCheck: true }
  }))
  expect(check(root).status).toBe(1)
})
