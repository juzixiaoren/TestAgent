# FlexMD 使用说明手册

FlexMD 是一个小型的、基于 Markdown 的幻灯片 DSL。它的目标编译命令是：

```bash
python -m flexmd compile input.fmd -o output.html --ast output.ast.json
```

测试智能体应把 FlexMD 视为一个**解析器/编译器类项目**，而不是普通的学生管理系统、CRUD 系统或业务系统。

## 一、核心语法

### 1. 幻灯片分页

在根层级，单独一行 `---` 表示分页。

```fmd
# Slide 1
---
# Slide 2
```

连续的根层级分页符是合法的。连续 `---` 会生成空 slide。

```fmd
# Slide 1
---
---
# Slide 3
```

预期 slide 数量：3。

FlexMD **不支持 Markdown 原生水平线**。`---` 被保留用于 slide 或 pane 分隔。

### 2. Flex 布局块

Flex block 用来定义布局容器。

```fmd
\h
left
---
right
\\h
```

`\h` 表示 horizontal layout，即横向布局，pane 从左到右排列。

```fmd
\v
top
---
bottom
\\v
```

`\v` 表示 vertical layout，即纵向布局，pane 从上到下排列。

在 flex block 内部，单独一行 `---` 表示 pane 分隔。

flex 内部连续 `---` 是合法的，会生成空 pane。

```fmd
\h
left
---
---
right
\\h
```

预期 pane 数量：3。

空 pane 是合法的。

### 3. 嵌套 flex

Flex block 可以嵌套。

````fmd
\v
\h
left
---
right
\\h
---
```python
print("hello")
```
\\v
````

### 4. 比例语法

flex 开始标记可以带可选 pane 比例。

```fmd
\h 1:2
left
---
right
\\h
```

对于横向布局，比例控制 pane 的相对宽度。对于纵向布局，比例控制 pane 的相对高度。

规则：

1. 比例语法是可选的。
2. 比例必须是正数，用 `:` 分隔。
3. 如果提供比例，比例数量必须与最终 pane 数量一致。
4. 非法比例应产生语法错误。
5. AST 应在 flex 节点上保留比例字段，字段名为 `ratios`，例如 `[1, 2]`。
6. HTML 应通过 flex 样式体现比例，例如 `flex-grow: 1` 和 `flex-grow: 2`。

### 5. 代码块

在 fenced code block 内部，FlexMD 特殊标记都应被视为普通文本。

下面这个输入只有一个 slide，而不是两个：

````fmd
# Demo
```python
print("before")
---
print("after")
```
````

下面这个输入不包含 flex block：

````fmd
```text
\h
---
\\h
```
````

### 6. 错误处理

编译器应拒绝以下输入：

- 未闭合的 flex block；
- flex 闭合标记不匹配，例如 `\h ... \\v`；
- 根层级出现多余闭合标记；
- 未闭合的 fenced code block；
- 非法比例，例如 `\h 1:0`、`\h 1:x`，或者 `\h 1:2` 实际有三个 pane。

## 二、HTML 渲染要求

输出应是一个单文件 HTML 幻灯片页面。

必要 CSS 行为：

- slide 应以演示页面形式显示；
- flex 容器应使用 CSS flex 布局；
- `\h` 应渲染为 `flex-direction: row`；
- `\v` 应渲染为 `flex-direction: column`；
- pane 比例应影响 flex growth；
- 生成的演示页面应隐藏滚动条。

隐藏滚动条可以使用类似 CSS：

```css
html, body, .deck, .slide, .pane {
  scrollbar-width: none;
}
html::-webkit-scrollbar,
body::-webkit-scrollbar,
.deck::-webkit-scrollbar,
.slide::-webkit-scrollbar,
.pane::-webkit-scrollbar {
  display: none;
}
```

## 三、推荐测试智能体流程

测试智能体应直接完成从用例设计到结果分析的闭环：

1. 读取本规格和 FlexMD 源码，明确测试目标。
2. 在执行前写明每个测试用例的测试目的和预期行为。
3. 使用工具实际写入 `.fmd` 测试文件。
4. 使用命令行将 `.fmd` 文件送入 FlexMD 编译器：

```bash
python -m flexmd compile input.fmd -o output.html --ast output.ast.json
```

5. 读取编译生成的 AST JSON 和 HTML 文件。
6. 根据本规格判断实际结果是否符合预期。
7. 对失败用例继续构造最小复现或相邻验证用例。
8. 生成中文测试报告，说明测试计划、执行命令、实际输出、失败原因和修改建议。

测试智能体不应只停留在口头分析，也不应只读取预生成报告；它需要实际编写测试输入、执行被测程序并分析结果。
