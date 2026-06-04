# Plan 41：前端路由现代化与多文件解耦重构

## 一、这一阶段解决了什么问题

在 Plan 40 之前，GUI 功能已经很完整，但前端应用层还有两个结构性问题：

1. 顶层 `frontend/src/App.tsx` 同时负责主题、导航、数据状态、后端调用、页面切换和布局渲染
2. 页面切换仍然依赖本地 `tab` 状态，而不是现代路由系统
3. 路由、布局、状态管理和页面容器没有明确边界
4. 随着控制台、网络页、交易页继续增长，顶层文件会越来越难维护

所以 Plan 41 的核心目标不是增加新业务功能，而是把当前 GUI 从“单文件堆叠应用”升级为“可维护、可扩展、可讲解的现代前端结构”。

## 二、为什么这一阶段必须做

这个项目虽然已经接近答辩基线，但前端如果继续保持单文件集中管理，会带来几个明显风险：

1. 小改动容易牵动整个 `App.tsx`
2. 路由层和业务层耦合，后续定位问题成本高
3. 很难向老师清楚解释前端分层
4. 一旦 GUI 再出现白屏或初始化问题，排查入口会非常混乱

因此这一步的价值在于“稳结构”，不是“堆功能”。

## 三、本阶段实现范围

### 已完成内容

1. 引入最新版 `@tanstack/react-router`
2. 引入 TanStack Router Vite 插件，改为文件路由生成方式
3. 把 Vite 配置改成 TypeScript 文件 `vite.config.ts`
4. 增加 `@` 与 `@wails` 路径别名，减少深层相对路径
5. 把主题切换抽到独立 Provider
6. 把应用数据与动作抽到 `WorkbenchProvider`
7. 把布局拆成独立文件：顶部栏、左侧导航、整体壳层
8. 把各功能页改成独立路由入口文件
9. 保留原有页面组件，使用“路由页容器 + 业务页组件”的方式降低改造风险
10. 同步补充本阶段计划文档

### 本阶段没有做的内容

1. 没有修改 Go 后端接口语义
2. 没有新增链功能、交易功能或网络功能
3. 没有重写现有各业务页 UI 细节
4. 没有引入 Zustand、Redux 之类的新状态库

## 四、核心理论讲解

### 1. 为什么这里要换成路由，而不是继续用 Tabs

本地 `tab` 状态只适合非常小的单页切换。
一旦应用已经有：

1. Dashboard
2. 钱包
3. 区块
4. 交易与挖矿
5. 网络
6. 控制台

这些功能时，本质上已经是一个“多工作区桌面客户端”，而不是一个单页里的几块内容。

用现代路由的好处是：

1. 页面边界清晰
2. 导航状态天然可追踪
3. 每个页面有独立入口
4. 更容易继续拆分 loader、layout、error boundary

### 2. 为什么选择 TanStack Router

当前前端已经是：

1. React 19
2. TypeScript 6
3. Vite 8

TanStack Router 在这套技术栈里很合适，因为它强调：

1. 强类型路由
2. 文件路由生成
3. 现代 React 生态兼容性好
4. 后续如果要继续做 loader、search params、嵌套路由，也能自然扩展

### 3. 为什么不是直接把所有页面都重写

如果为了“换路由”顺便把所有页面重新实现，会让一次改动同时承担：

1. 路由替换风险
2. 页面重写风险
3. UI 回归风险
4. 状态回归风险

这不适合当前答辩阶段。

所以这里采用更稳的方案：

1. 先保留现有业务页组件
2. 新增一层路由页容器
3. 把数据和动作从大 `App.tsx` 中抽出来
4. 再由容器把数据传给原页面组件

这样既完成了结构升级，又把行为变化控制在最小范围。

### 4. 为什么要增加 Provider 分层

原来的 `App.tsx` 同时承担：

1. 主题状态
2. 页面切换状态
3. 数据刷新逻辑
4. 命令执行逻辑
5. 网络操作逻辑
6. 布局渲染

这会导致一个文件承担过多职责。

现在拆成两类 Provider：

1. `ThemeModeProvider`：只负责深浅色与 MUI 主题
2. `WorkbenchProvider`：只负责桌面客户端的数据、表单和动作

这样分层以后，每个模块更容易解释和维护。

## 五、代码结构讲解

本阶段新增或重构的核心目录如下：

1. `frontend/src/app/`
2. `frontend/src/features/workbench/`
3. `frontend/src/routes/`
4. `frontend/src/App.tsx`
5. `frontend/vite.config.ts`
6. `frontend/tsconfig.json`
7. `docs/plan/plan41.md`

### 1. `frontend/src/app/`

这一层放应用级基础设施：

1. `router.tsx`：创建 TanStack Router 实例
2. `layout/AppShell.tsx`：整体桌面壳层
3. `layout/AppHeader.tsx`：顶部工具栏
4. `layout/AppSidebar.tsx`：左侧导航
5. `navigation/`：导航路径和菜单项定义
6. `providers/ThemeModeProvider.tsx`：深浅色模式 Provider

这部分不放具体业务逻辑，目的是保证“应用壳”和“业务工作台”分离。

### 2. `frontend/src/features/workbench/`

这一层放当前桌面工作台的业务状态与动作：

1. `hooks/useWorkbenchController.ts`：把原先大 `App.tsx` 中的状态与事件逻辑抽离出来
2. `context/WorkbenchProvider.tsx`：向全应用暴露工作台上下文
3. `context/useWorkbench.ts`：业务访问入口
4. `route-pages/`：每个路由页对应一个轻量容器组件
5. `types.ts`：集中定义工作台表单与上下文类型

### 3. `frontend/src/routes/`

这是真正的文件路由目录：

1. `__root.tsx`：根布局路由
2. `index.tsx`：Dashboard
3. `wallets.tsx`
4. `blocks.tsx`
5. `transactions.tsx`
6. `network.tsx`
7. `console.tsx`

TanStack Router 插件会根据这些文件生成 `routeTree.gen.ts`。

### 4. `frontend/src/App.tsx`

现在顶层只负责三件事：

1. 注入主题 Provider
2. 注入业务 Provider
3. 注入 RouterProvider

这说明顶层文件已经恢复成“应用装配层”，而不是“超级业务文件”。

## 六、实现步骤复盘

### 第一步：确认当前 GUI 的职责边界

先识别原 `App.tsx` 同时承担的所有职责，明确哪些应该归到路由、布局、Provider 和业务控制器中。

### 第二步：安装最新版 TanStack Router

使用 Bun 直接安装最新版本的：

1. `@tanstack/react-router`
2. `@tanstack/router-plugin`

这样可以保证当前项目接入的是当前时间点的最新主线版本。

### 第三步：改造构建配置

把 `vite.config.js` 升级为 `vite.config.ts`，同时接入：

1. React Compiler 插件
2. TanStack Router 插件
3. 路径别名配置

### 第四步：拆应用壳

把顶部栏、左侧导航、整体布局从原始 `App.tsx` 中拆出来，形成稳定的桌面客户端壳层。

### 第五步：抽业务控制器

把原来集中在 `App.tsx` 的：

1. 刷新逻辑
2. 钱包创建
3. 主链初始化
4. 交易提交
5. 多签花费
6. 挖矿
7. CLI 执行
8. 网络节点操作
9. 网络演示流程

统一迁移到 `useWorkbenchController.ts`。

### 第六步：补路由入口层

为每个页面增加单独的路由文件与容器页，使页面切换完全交给 TanStack Router。

### 第七步：生成路由树并做构建验证

通过前端构建生成 `routeTree.gen.ts`，再执行类型检查与完整构建验证。

## 七、关键代码讲解

### 1. `ThemeModeProvider`

这个 Provider 只做一件事：

1. 维护 `light / dark`
2. 通过 `createTheme({ palette: { mode } })` 保持 MUI 默认主题语义
3. 把模式写入 `localStorage`

这满足了“需要深浅色切换”与“尽量使用 MUI 默认风格”这两个目标。

### 2. `useWorkbenchController`

这是这次重构里最重要的业务控制层。
它把原先顶层组件中的大量 `useState + handler` 全部转移到专门的 hook 里。

这样做的结果是：

1. 业务状态集中
2. 顶层布局不再理解具体业务细节
3. 页面容器只做连接，不做真正业务实现

### 3. 路由页容器

像 `DashboardRoutePage.tsx`、`NetworkRoutePage.tsx` 这类文件只负责：

1. 从 `useWorkbench()` 取数据
2. 把数据转交给已有页面组件

这层的意义是“削峰填谷”：

1. 上游对接路由
2. 下游复用既有页面
3. 中间不增加额外业务复杂度

### 4. 文件路由目录

文件路由的好处在于目录结构本身就是路由结构。
即使不打开 router 配置文件，看到 `src/routes/` 也能立刻理解当前页面体系。

## 八、常见错误

### 错误 1：换了路由，但仍把所有状态继续堆在顶层

这样只是“表面用了 Router”，本质上还是单文件应用。

### 错误 2：为了换路由顺手重写全部页面

这会让改动过大，很难验证回归风险。

### 错误 3：没有把布局层和业务层分开

这样后续任何导航改动都可能牵动业务代码。

### 错误 4：继续使用深层相对路径

随着文件变多，`../../../../` 这种导入会迅速降低可维护性，所以本阶段同步引入了路径别名。

## 九、验收结果

本阶段完成后，应满足：

1. GUI 已使用 TanStack Router 管理页面切换
2. 左侧导航切换不再依赖本地 Tabs 状态
3. 顶层 `App.tsx` 已大幅瘦身
4. 主题、布局、业务控制器、路由入口已经拆分
5. 前端验证通过：
   - `bunx tsc --noEmit`
   - `bun run build`
6. 整体桌面验证通过：
   - `wails build -clean`

## 十、答辩时可以怎么讲

这一阶段可以这样介绍：

1. 之前 GUI 功能虽然完整，但顶层结构太集中
2. 如果继续迭代，会让维护和讲解都越来越困难
3. 所以这里没有继续加新功能，而是把前端结构升级为现代路由架构
4. 我引入了 TanStack Router，把页面、布局、状态和主题分层拆开
5. 这样后续无论是继续维护，还是在答辩中解释整体前端结构，都更清晰

## 十一、当前阶段结论

Plan 41 完成后，GUI 的前端组织方式已经比之前更接近一个真正可维护的桌面客户端：

1. 路由清晰
2. 布局独立
3. 业务状态集中
4. 主题能力独立
5. 多文件结构更适合继续维护

这一阶段属于“结构收口型重构”，对答辩质量和后续稳定性都有直接帮助。
