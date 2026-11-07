import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'
import { parse } from 'vue/compiler-sfc'

const root = path.resolve(process.argv[2] || fileURLToPath(new URL('..', import.meta.url)))
const problems = []
const configPath = path.join(root, 'tsconfig.app.json')
const configuration = ts.readConfigFile(configPath, ts.sys.readFile)
if (configuration.error) throw new Error(ts.flattenDiagnosticMessageText(configuration.error.messageText, '\n'))
const config = configuration.config
const parsed = ts.parseJsonConfigFileContent(config, ts.sys, root, undefined, configPath, undefined,
  [{ extension: '.vue', isMixedContent: true, scriptKind: ts.ScriptKind.Deferred }])
for (const error of parsed.errors) problems.push(ts.flattenDiagnosticMessageText(error.messageText, '\n'))
const options = parsed.options
function checkOptions(options, name) {
  if (options.strict !== true || options.allowJs !== false || options.noCheck === true) {
    problems.push(`${name}: strict:true and allowJs:false are required; noCheck must not disable checking`)
  }
  for (const flag of ['noImplicitAny', 'noImplicitThis', 'strictNullChecks', 'strictFunctionTypes',
    'strictBindCallApply', 'strictPropertyInitialization', 'strictBuiltinIteratorReturn', 'useUnknownInCatchVariables']) {
    if (options[flag] === false) problems.push(`${name}: strict sub-option ${flag} must not be disabled`)
  }
  if (options.baseUrl !== undefined || options.ignoreDeprecations !== undefined) {
    problems.push(`${name}: deprecated baseUrl and ignoreDeprecations must not hide migration errors`)
  }
}
checkOptions(options, 'tsconfig.app.json')
for (const name of ['tsconfig.node.json', 'tsconfig.test.json']) {
  const file = path.join(root, name)
  if (!fs.existsSync(file)) continue
  const result = ts.readConfigFile(file, ts.sys.readFile)
  if (result.error) { problems.push(ts.flattenDiagnosticMessageText(result.error.messageText, '\n')); continue }
  const project = ts.parseJsonConfigFileContent(result.config, ts.sys, root, undefined, file, undefined,
    [{ extension: '.vue', isMixedContent: true, scriptKind: ts.ScriptKind.Deferred }])
  for (const error of project.errors) problems.push(ts.flattenDiagnosticMessageText(error.messageText, '\n'))
  checkOptions(project.options, name)
}
for (const pattern of ['src/**/*.ts', 'src/**/*.vue']) {
  if (!config.include?.includes(pattern)) problems.push(`include must cover ${pattern}`)
}
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
