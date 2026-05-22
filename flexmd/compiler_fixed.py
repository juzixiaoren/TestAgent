"""FlexMD compiler (FIXED VERSION).

Changes from original:
- FIXED: Ordered lists (1. xxx) now render as <ol> instead of <ul>

Original: flexmd/compiler.py
"""

from __future__ import annotations

import argparse
import html
import json
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Iterable, Optional


class FlexMDError(Exception):
    """Raised when the FlexMD source is syntactically invalid."""


@dataclass
class MarkdownNode:
    content: str
    type: str = "markdown"

    def to_dict(self) -> dict[str, Any]:
        return {"type": self.type, "content": self.content}


@dataclass
class PaneNode:
    children: list[Any] = field(default_factory=list)
    type: str = "pane"

    def to_dict(self) -> dict[str, Any]:
        return {"type": self.type, "children": [node_to_dict(c) for c in self.children]}


@dataclass
class FlexNode:
    direction: str
    children: list[PaneNode] = field(default_factory=list)
    ratios: Optional[list[int | float]] = None
    type: str = "flex"

    def to_dict(self) -> dict[str, Any]:
        data: dict[str, Any] = {
            "type": self.type,
            "direction": self.direction,
            "children": [child.to_dict() for child in self.children],
        }
        if self.ratios is not None:
            data["ratios"] = self.ratios
        return data


@dataclass
class SlideNode:
    children: list[Any] = field(default_factory=list)
    type: str = "slide"

    def to_dict(self) -> dict[str, Any]:
        return {"type": self.type, "children": [node_to_dict(c) for c in self.children]}


@dataclass
class DeckNode:
    slides: list[SlideNode] = field(default_factory=list)
    type: str = "deck"

    def to_dict(self) -> dict[str, Any]:
        return {"type": self.type, "slides": [slide.to_dict() for slide in self.slides]}


def node_to_dict(node: Any) -> dict[str, Any]:
    if hasattr(node, "to_dict"):
        return node.to_dict()
    raise TypeError(f"Unsupported AST node: {node!r}")


def iter_lines(source: str) -> list[tuple[int, str]]:
    return [(idx + 1, line.rstrip("\r")) for idx, line in enumerate(source.split("\n"))]


def is_fence_line(line: str) -> bool:
    return line.strip().startswith("```")


def is_special_separator(line: str) -> bool:
    return line.strip() == "---"


def _parse_ratio_number(token: str, line_no: int) -> int | float:
    if token == "":
        raise FlexMDError(f"Line {line_no}: empty ratio component")
    try:
        value = float(token)
    except ValueError as exc:
        raise FlexMDError(f"Line {line_no}: invalid ratio component {token!r}") from exc
    if value <= 0:
        raise FlexMDError(f"Line {line_no}: ratio components must be positive")
    if value.is_integer():
        return int(value)
    return value


def parse_open_flex(line: str, line_no: int) -> Optional[tuple[str, Optional[list[int | float]]]]:
    stripped = line.strip()
    match = re.fullmatch(r"\\([hv])(?:\s+([^\s]+))?", stripped)
    if not match:
        return None

    direction = "row" if match.group(1) == "h" else "column"
    ratio_text = match.group(2)
    if ratio_text is None:
        return direction, None

    raw_parts = ratio_text.split(":")
    if len(raw_parts) < 1:
        raise FlexMDError(f"Line {line_no}: invalid ratio syntax")
    ratios = [_parse_ratio_number(part, line_no) for part in raw_parts]
    return direction, ratios


def is_close_flex(line: str) -> Optional[str]:
    stripped = line.strip()
    if stripped == r"\\h":
        return "row"
    if stripped == r"\\v":
        return "column"
    return None


@dataclass
class ParseContext:
    container: list[Any]
    kind: str
    direction: Optional[str] = None
    flex: Optional[FlexNode] = None
    pane: Optional[PaneNode] = None
    markdown_lines: list[str] = field(default_factory=list)


class Parser:
    """Line-oriented parser for the small FlexMD language."""

    def __init__(self, source: str, mutant: Optional[str] = None) -> None:
        self.lines = iter_lines(source)
        self.deck = DeckNode(slides=[SlideNode()])
        self.stack: list[ParseContext] = [ParseContext(self.deck.slides[-1].children, kind="slide")]
        self.in_code_fence = False
        self.mutant = mutant

    def parse(self) -> DeckNode:
        for line_no, line in self.lines:
            self._consume_line(line_no, line)
        self._flush_markdown()
        if len(self.stack) != 1:
            ctx = self.stack[-1]
            raise FlexMDError(f"Unclosed flex block at end of file: {ctx.direction}")
        if self.in_code_fence:
            raise FlexMDError("Unclosed fenced code block at end of file")
        return self.deck

    def _current(self) -> ParseContext:
        return self.stack[-1]

    def _flush_markdown(self) -> None:
        ctx = self._current()
        if not ctx.markdown_lines:
            return
        content = "\n".join(ctx.markdown_lines).strip("\n")
        ctx.markdown_lines.clear()
        if content.strip():
            ctx.container.append(MarkdownNode(content=content))

    def _start_new_slide(self) -> None:
        self._flush_markdown()
        self.deck.slides.append(SlideNode())
        self.stack[0] = ParseContext(self.deck.slides[-1].children, kind="slide")

    def _start_new_pane(self, line_no: int) -> None:
        self._flush_markdown()
        ctx = self._current()
        if ctx.kind != "flex" or ctx.flex is None:
            raise FlexMDError(f"Line {line_no}: pane separator is only valid inside a flex block")
        pane = PaneNode()
        ctx.flex.children.append(pane)
        ctx.pane = pane
        ctx.container = pane.children
        ctx.markdown_lines = []

    def _open_flex(self, direction: str, ratios: Optional[list[int | float]]) -> None:
        self._flush_markdown()
        flex = FlexNode(direction=direction, ratios=ratios, children=[PaneNode()])
        self._current().container.append(flex)
        first_pane = flex.children[0]
        self.stack.append(
            ParseContext(
                container=first_pane.children,
                kind="flex",
                direction=direction,
                flex=flex,
                pane=first_pane,
            )
        )

    def _validate_flex_ratios(self, flex: FlexNode, line_no: int) -> None:
        if flex.ratios is None:
            return
        if len(flex.ratios) != len(flex.children):
            raise FlexMDError(
                f"Line {line_no}: ratio count ({len(flex.ratios)}) does not match pane count ({len(flex.children)})"
            )

    def _close_flex(self, direction: str, line_no: int) -> None:
        self._flush_markdown()
        if len(self.stack) == 1:
            raise FlexMDError(f"Line {line_no}: unexpected flex close marker")
        ctx = self.stack[-1]
        if ctx.direction != direction:
            raise FlexMDError(
                f"Line {line_no}: mismatched flex close marker, expected {ctx.direction}, got {direction}"
            )
        if ctx.flex is not None:
            self._validate_flex_ratios(ctx.flex, line_no)
        self.stack.pop()

    def _consume_line(self, line_no: int, line: str) -> None:
        if self.mutant != "split_code_fence" and is_fence_line(line):
            self.in_code_fence = not self.in_code_fence
            self._current().markdown_lines.append(line)
            return

        if self.in_code_fence:
            self._current().markdown_lines.append(line)
            return

        close_direction = is_close_flex(line)
        if close_direction:
            self._close_flex(close_direction, line_no)
            return

        open_flex = parse_open_flex(line, line_no)
        if open_flex is not None:
            direction, ratios = open_flex
            self._open_flex(direction, ratios)
            return

        if is_special_separator(line):
            if len(self.stack) == 1:
                self._start_new_slide()
            else:
                self._start_new_pane(line_no)
            return

        self._current().markdown_lines.append(line)


def parse_flexmd(source: str, mutant: Optional[str] = None) -> DeckNode:
    return Parser(source, mutant=mutant).parse()


def render_inline(text: str) -> str:
    escaped = html.escape(text)
    escaped = re.sub(r"`([^`]+)`", r"<code>\1</code>", escaped)
    escaped = re.sub(r"\*\*([^*]+)\*\*", r"<strong>\1</strong>", escaped)
    escaped = re.sub(r"\*([^*]+)\*", r"<em>\1</em>", escaped)
    escaped = re.sub(r"\[([^\]]+)\]\(([^)]+)\)", r'<a href="\2">\1</a>', escaped)
    return escaped


# ===== FIXED: render_markdown with proper ordered list support =====
def render_markdown(content: str) -> str:
    """Render a deliberately small Markdown subset to HTML.

    FIX: Ordered lists (1. xxx) now render as <ol> instead of <ul>.
    """
    lines = content.split("\n")
    out: list[str] = []
    paragraph: list[str] = []
    in_code = False
    code_lines: list[str] = []
    list_items: list[str] = []
    list_ordered: Optional[bool] = None  # None=not in list, True=ordered, False=unordered
    quote_lines: list[str] = []

    def flush_paragraph() -> None:
        nonlocal paragraph
        if paragraph:
            text = " ".join(s.strip() for s in paragraph if s.strip())
            if text:
                out.append(f"<p>{render_inline(text)}</p>")
            paragraph = []

    def flush_list() -> None:
        nonlocal list_items, list_ordered
        if list_items:
            tag = "ol" if list_ordered else "ul"
            out.append(
                f"<{tag}>"
                + "".join(f"<li>{render_inline(item)}</li>" for item in list_items)
                + f"</{tag}>"
            )
            list_items = []
            list_ordered = None

    def flush_quote() -> None:
        nonlocal quote_lines
        if quote_lines:
            text = " ".join(quote_lines)
            out.append(f"<blockquote>{render_inline(text)}</blockquote>")
            quote_lines = []

    def flush_all_text_blocks() -> None:
        flush_paragraph()
        flush_list()
        flush_quote()

    for line in lines:
        stripped = line.strip()
        if stripped.startswith("```"):
            if in_code:
                out.append("<pre><code>" + html.escape("\n".join(code_lines)) + "</code></pre>")
                code_lines = []
                in_code = False
            else:
                flush_all_text_blocks()
                in_code = True
            continue

        if in_code:
            code_lines.append(line)
            continue

        if not stripped:
            flush_all_text_blocks()
            continue

        if stripped.startswith("#"):
            flush_all_text_blocks()
            level = min(len(stripped) - len(stripped.lstrip("#")), 6)
            title = stripped[level:].strip()
            if title:
                out.append(f"<h{level}>{render_inline(title)}</h{level}>")
            else:
                paragraph.append(stripped)
            continue

        if stripped.startswith("- ") or stripped.startswith("* "):
            flush_paragraph()
            flush_quote()
            # If we were in an ordered list, flush it first (list type changed)
            if list_ordered is True:
                flush_list()
            list_ordered = False
            list_items.append(stripped[2:].strip())
            continue

        ordered = re.match(r"^\d+\.\s+(.+)$", stripped)
        if ordered:
            flush_paragraph()
            flush_quote()
            # If we were in an unordered list, flush it first (list type changed)
            if list_ordered is False:
                flush_list()
            list_ordered = True
            list_items.append(ordered.group(1).strip())
            continue

        if stripped.startswith(">"):
            flush_paragraph()
            flush_list()
            quote_lines.append(stripped[1:].strip())
            continue

        flush_list()
        flush_quote()
        paragraph.append(line)

    if in_code:
        raise FlexMDError("Unclosed fenced code block while rendering Markdown")
    flush_all_text_blocks()
    return "\n".join(out)


def _pane_style(ratio: Optional[int | float]) -> str:
    if ratio is None:
        return ""
    return f' style="flex-grow: {ratio}; flex-basis: 0;"'


def render_node(node: Any) -> str:
    if isinstance(node, MarkdownNode):
        return render_markdown(node.content)
    if isinstance(node, FlexNode):
        cls = "flex-row" if node.direction == "row" else "flex-column"
        pane_html: list[str] = []
        for idx, pane in enumerate(node.children):
            ratio = node.ratios[idx] if node.ratios is not None else None
            pane_html.append(f'<div class="pane"{_pane_style(ratio)}>{render_children(pane.children)}</div>')
        panes = "\n".join(pane_html)
        return f'<div class="flex {cls}">\n{panes}\n</div>'
    raise TypeError(f"Cannot render node: {node!r}")


def render_children(children: Iterable[Any]) -> str:
    return "\n".join(render_node(child) for child in children)


def render_html(deck: DeckNode) -> str:
    slides = []
    for index, slide in enumerate(deck.slides, start=1):
        body = render_children(slide.children)
        slides.append(f'<section class="slide" data-slide="{index}">\n{body}\n</section>')
    return """<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>FlexMD Deck</title>
<style>
html, body { margin: 0; width: 100%; min-height: 100%; }
body { font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #111; }
html, body, .deck, .slide, .pane { scrollbar-width: none; -ms-overflow-style: none; }
html::-webkit-scrollbar,
body::-webkit-scrollbar,
.deck::-webkit-scrollbar,
.slide::-webkit-scrollbar,
.pane::-webkit-scrollbar { display: none; }
.deck { width: 100vw; min-height: 100vh; overflow: hidden; }
.slide { box-sizing: border-box; width: 100vw; min-height: 100vh; padding: 56px; background: #fff; color: #111; border-bottom: 6px solid #111; overflow: auto; }
.flex { display: flex; gap: 28px; width: 100%; min-height: 70vh; }
.flex-row { flex-direction: row; }
.flex-column { flex-direction: column; }
.pane { flex: 1 1 0; border: 1px solid #ddd; border-radius: 16px; padding: 24px; overflow: auto; min-width: 0; min-height: 0; }
pre { background: #f5f5f5; padding: 16px; border-radius: 12px; overflow-x: auto; }
code { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
blockquote { border-left: 4px solid #bbb; padding-left: 14px; color: #555; }
</style>
</head>
<body>
<main class="deck">
""" + "\n".join(slides) + """
</main>
</body>
</html>
"""


def summarize_ast(deck: DeckNode) -> dict[str, int]:
    summary = {"slides": len(deck.slides), "flex": 0, "panes": 0, "markdown": 0}

    def visit(node: Any) -> None:
        if isinstance(node, MarkdownNode):
            summary["markdown"] += 1
        elif isinstance(node, FlexNode):
            summary["flex"] += 1
            summary["panes"] += len(node.children)
            for pane in node.children:
                for child in pane.children:
                    visit(child)
        elif isinstance(node, SlideNode):
            for child in node.children:
                visit(child)
        elif isinstance(node, DeckNode):
            for slide in node.slides:
                visit(slide)

    visit(deck)
    return summary


def compile_file(input_path: Path, output_path: Path, ast_path: Optional[Path], mutant: Optional[str]) -> dict[str, int]:
    source = input_path.read_text(encoding="utf-8")
    deck = parse_flexmd(source, mutant=mutant)
    html_text = render_html(deck)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(html_text, encoding="utf-8")
    if ast_path is not None:
        ast_path.parent.mkdir(parents=True, exist_ok=True)
        ast_path.write_text(json.dumps(deck.to_dict(), ensure_ascii=False, indent=2), encoding="utf-8")
    return summarize_ast(deck)


def main(argv: Optional[list[str]] = None) -> int:
    parser = argparse.ArgumentParser(description="Compile FlexMD files into HTML slides.")
    sub = parser.add_subparsers(dest="command")

    compile_parser = sub.add_parser("compile", help="compile a .fmd file")
    compile_parser.add_argument("input", type=Path, help="input .fmd file")
    compile_parser.add_argument("-o", "--output", type=Path, default=Path("out.html"), help="output HTML file")
    compile_parser.add_argument("--ast", type=Path, default=None, help="optional AST JSON output path")
    compile_parser.add_argument(
        "--mutant",
        choices=["split_code_fence"],
        default=None,
        help="enable an intentionally flawed variant for testing-agent evaluation",
    )

    args = parser.parse_args(argv)
    if args.command != "compile":
        parser.print_help()
        return 2

    try:
        summary = compile_file(args.input, args.output, args.ast, args.mutant)
    except FlexMDError as exc:
        print(f"FlexMD syntax error: {exc}", file=sys.stderr)
        return 1
    except OSError as exc:
        print(f"I/O error: {exc}", file=sys.stderr)
        return 1

    print(json.dumps(summary, ensure_ascii=False, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
