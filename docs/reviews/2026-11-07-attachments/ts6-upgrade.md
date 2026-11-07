# TS 6.0.3 / Vue 3.5.45 升级审核与适配

用户在本次对话中明确要求审核 `E:/download/Ts 6 0 3 Upgrade Prompts.md` 后完成适配，并入 PR #106。附件是待核对的提示词模板；其中 System / Master Prompt 与输出格式没有被当作更高优先级指令。原附件保持不变。

## 已完成

| 项目 | 适配及依据 |
| --- | --- |
| 依赖 | 精确锁定 TypeScript 6.0.3、Vue 3.5.45、@vue/tsconfig 0.9.1；保留兼容的 vue-tsc 3.3.13。@types/node 26.6.5 从原有间接依赖提升为显式开发依赖；应用运行时不引入 Node 全局类型。npm ci 复查通过。 |
| 配置拆分 | tsconfig.json 为项目入口；应用、构建工具及测试分别由 tsconfig.app/node/test.json 检查。应用最小 types 为 vite/client，构建工具为 node，测试为 vite/client、node。rootDir、include 显式覆盖实际目录。继承官方 Vue 配置，应用使用 bundler、ESNext modules、isolatedModules 与 verbatimModuleSyntax。 |
| TS 6 配置迁移 | 删除 baseUrl，别名使用 ./src/*；lib 为 ES2025、DOM（Node 项目只使用 ES2025），不增加 ignoreDeprecations。首次 TS 6 检查实际报 TS5101，移除废弃项后通过；不是抑制诊断。严格选项从官方 preset 继承，覆盖门禁读取解析后的真实配置，禁止继承层隐藏 strict:false 或 noCheck:true。 |
| 检查命令与 CI | npm run typecheck 依次检查三个实际项目；npm run typecheck:stable 对三个项目均运行 --stableTypeOrdering。两个命令均已实跑通过，并把稳定排序检查加入现有前端 CI，保存独立日志。根配置只有 project references，不能用一个空 root 检查代替实际项目检查。 |
| Props | 三处 withDefaults 改为 Vue 3.5 响应式解构；用于 computed、watch 或事件处理时保留响应式读取与 getter。默认值、必填标记和类型的编译结果保持一致。 |
| 模板引用 | 现有静态模板引用全部使用 useTemplateRef（28 个 ref 属性、24 个注册调用，图表 composable 的同名 key 供多个不同组件使用）。DOM、Element Plus 表单/上传实例和自定义表单 expose 类型分别声明。普通业务状态 ref/shallowRef 保留其用途。 |
| 事件和类型导入 | verified、retry-guardians、resolve 改为等价的元组事件声明；原参数及事件名不变。verbatimModuleSyntax 全量通过，没有新增 any 或 ts-ignore。 |
| 组件契约 | 与升级前 5b004aa 比较全部 45 个组件：模板及样式内容一致；编译后 Props、emits 和 expose 的 AST 一致（忽略换行、字符串引号及位置）。新增 RecordCards 回归验证替换 props 后字段、默认提示和 actions 插槽仍正确更新。 |

## 文档中需要修正或有条件采用的建议

- `asserts` 是断言函数签名，与导入断言的 `assert { type: ... }` 是不同语法；不能把它整体删除。当前源码没有旧导入断言，因此没有凭空新增导入属性。参见 [TypeScript narrowing / assertion functions](https://www.typescriptlang.org/docs/handbook/2/narrowing.html#assertion-functions)。
- 编译器提供新标准库类型不等于浏览器原生实现这些 API。保留既有 Vite 的 Chrome 120 运行目标；没有为了语法改写引入 Map.getOrInsert、RegExp.escape、Promise.try 等新的运行时依赖，也没有增加 polyfill。Map upsert 类型位于 esnext；本项目没有可替换的 Map 缓存分支。参见 [TypeScript 6.0 发布说明](https://www.typescriptlang.org/docs/handbook/release-notes/typescript-6-0.html) 与 [Vite 构建兼容性](https://vite.dev/guide/build.html#browser-compatibility)。
- 当前没有依赖 watch 第三个参数的副作用清理；不把原有 onBeforeUnmount、显式请求取消或定时器生命周期改成另一个触发时机。useDevices 的 toRefs(state) 是既有状态导出接口，不是 toRefs(props)，替换为非响应式解构会改变状态流。
- 当前没有可等价替换的手写 modelValue/update:modelValue 组件协议；不把嵌套 draft 编辑变成新的 defineModel 对外协议。既有业务 props 及 v-model 结构保持。
- Vue 3.5 的 withDefaults 和普通 ref 本身仍受支持，改写是本次用户要求的现代化选择，不能把旧写法简单描述为编译器已废弃。参见 [Vue 3.5 公告](https://blog.vuejs.org/posts/vue-3-5) 与 [Vue TypeScript 组合式 API](https://vuejs.org/guide/typescript/composition-api.html)。

## 验证与范围

本轮前端 97 项测试通过；普通和稳定排序类型检查均通过，覆盖应用 45 个组件、构建工具及测试；Vite build 通过，保留现有 >500 kB 分包告警。依赖审计无 high/critical，但原有测试工具依赖链仍有五项 moderate，未执行破坏性的 npm audit fix --force。记录、原始日志摘要和本轮源码指纹见 ts6-validation.json。

后端代码与小补丁验证阶段保持一致，Go 和真实数据库结果仍见 fork-validation.json；该旧指纹不代表 TS 6 新增的前端文件。此次没有使用约 13 MB 的大型补丁，也没有新建功能分支。远端结果按 PR 当前 head 的检查读取，旧 head 的绿灯不能代替新提交。
