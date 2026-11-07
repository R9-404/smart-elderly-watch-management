import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'
import { parse } from 'vue/compiler-sfc'

const root = path.resolve(process.argv[2] || fileURLToPath(new URL('..', import.meta.url)))
const problems = []
const configuration = ts.readConfigFile(path.join(root, 'tsconfig.json'), ts.sys.readFile)
if (configuration.error) throw new Error(ts.flattenDiagnosticMessageText(configuration.error.messageText, '\n'))
const config = configuration.config
const options = config.compilerOptions || {}
if (options.strict !== true || options.allowJs !== false || options.noCheck === true) {
  problems.push('strict:true and allowJs:false are required; noCheck must not disable checking')
}
for (const flag of ['noImplicitAny', 'noImplicitThis', 'strictNullChecks', 'strictFunctionTypes',
  'strictBindCallApply', 'strictPropertyInitialization', 'strictBuiltinIteratorReturn', 'useUnknownInCatchVariables']) {
  if (options[flag] === false) problems.push(`strict sub-option ${flag} must not be disabled`)
}
for (const pattern of ['src/**/*.ts', 'src/**/*.vue']) {
  if (!config.include?.includes(pattern)) problems.push(`include must cover ${pattern}`)
}
if (config.exclude?.length || config.files?.length || config.extends) {
  problems.push('source coverage must remain explicit without files, exclude or inherited overrides')
}
const parsed = ts.parseJsonConfigFileContent(config, ts.sys, root, undefined, undefined, undefined,
  [{ extension: '.vue', isMixedContent: true, scriptKind: ts.ScriptKind.Deferred }])
const included = new Set(parsed.fileNames.map(file => path.resolve(file)))
let components = 0
function inspect(directory) {
  for (const item of fs.readdirSync(directory, { withFileTypes: true })) {
    const file = path.join(directory, item.name)
    if (item.isSymbolicLink()) { problems.push(`source symlink needs explicit scope review: ${file}`); continue }
    if (item.isDirectory()) { inspect(file); continue }
    if (/\.(?:[cm]?js|jsx)$/.test(item.name)) problems.push(`JavaScript source bypasses TS: ${file}`)
    if (/\.(?:[cm]?ts|tsx|vue)$/.test(item.name) && !included.has(path.resolve(file))) {
      problems.push(`source file is absent from the compiler include set: ${file}`)
    }
    if (!item.name.endsWith('.vue')) continue
    components++
    const { descriptor, errors } = parse(fs.readFileSync(file, 'utf8'), { filename: file })
    if (errors.length) problems.push(`invalid Vue component: ${file}`)
    for (const script of [descriptor.script, descriptor.scriptSetup].filter(Boolean)) {
      if (script.lang !== 'ts' || script.src) problems.push(`Vue script must be inline lang="ts": ${file}`)
    }
  }
}
inspect(path.join(root, 'src'))
if (problems.length) {
  process.stderr.write(problems.join('\n') + '\n')
  process.exitCode = 1
} else {
  process.stdout.write(`TypeScript scope verified: ${components} Vue components; full src coverage\n`)
}
