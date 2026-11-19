---
name: alt-ui-walk
description: Use for any UI change — building from a mock, reviewing a screen, a designer or qa walkthrough, or judging spacing, fonts or overlap. Sets the visual bar and the Playwright MCP walk.
---

# UI walk

## Building

- An owner-provided mock is the visual target: fonts, shell, spacing, tables, badges and hosted pages match it. Matching the structure alone is not enough.
- Slice the mock per screen, port its class lists, bundle its fonts. Start from primitives in Storybook.
- Read every note in the mock (pills marking existing, redesigned, new, needs backend). Notes decide intent; markup decides look.
- Each app follows its own mock screens. Apps share primitives and grammar only.
- Give builders an explicit box model in numbers, never "match the prototype".
- Confirm classes with the daisyUI skill and the daisyUI docs MCP; confirm framework APIs with context7.
- Never backlog an item the brief lists as frontend work without telling the owner.

## The bar

- Every box is symmetric: equal top and bottom, equal left and right space; rows and buttons vertically centered; equal gaps between siblings; buttons never wrap.
- Fonts match the mock exactly: computed family, size, weight, line height, letter spacing and icon sizes, side by side. The usual miss is base text where the size class is missing.
- Only data-dependent panels may be hidden.

## Walk (Playwright MCP)

1. Open each changed screen and state. English first; any other locale once at the end, then switch back.
2. Side-by-side screenshots against the mock at 1280 and 390, light and dark.
3. Final gate only: the overlap scan with `browser_evaluate` at 390, 768, 1024 and 1280, light and dark. It flags elements whose scrollWidth exceeds clientWidth (except intended scrollers), intersecting text rects, a document wider than the viewport, asymmetric boxes and wrapped controls. Zero hits before reporting.
4. Measure computed styles before a screenshot: a screenshot drops real `:hover`. Force states with CDP `CSS.forcePseudoState`.
5. Clear routes between phases with `page.unrouteAll({ behavior: 'ignoreErrors' })`.
6. Output goes in `.playwright-mcp/`. Call `browser_close` when done.

## Who walks

- A feature: pm, product, designer and qa each walk at review.
- A fix round or small UI change: one designer walk.
- The designer signs off and the lead looks at the screenshots before the owner sees anything.
