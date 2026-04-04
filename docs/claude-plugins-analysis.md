# Claude Code Plugins Analysis

## Current Setup (8 plugins)

| Plugin | Category |
|---|---|
| gopls-lsp | Go LSP |
| serena | Semantic code analysis |
| context7 | Documentation lookup |
| firecrawl | Web scraping/search |
| frontend-design | UI/UX |
| feature-dev | Feature development workflow |
| pr-review-toolkit | Code review |
| chrome-devtools-mcp | Browser debugging |

## Recommendations by Interest Area

### MCP Writing
- **`mcp-server-dev@claude-plugins-official`** - Skills specifically for designing and building MCP servers. Very relevant given the MCP backend (`backend/app/mcp/`).
- **`plugin-dev@claude-plugins-official`** - Plugin development toolkit (agents, hooks, MCP integrations). Useful for building plugins too.

### Go
- **gopls-lsp** - Already installed, this is the right one.
- **serena** - Already installed, gives semantic code nav on top of gopls.

### Documentation Querying
- **context7** - Already installed. Covers library docs lookup.
- **`claude-md-management@claude-plugins-official`** - Audit/improve CLAUDE.md files, capture session learnings. Could help keep the already-detailed CLAUDE.md current.

### Frontend Design
- **frontend-design** - Already installed.
- **chrome-devtools-mcp** - Already installed. Good for WASM debugging.

### Git / Monorepo / Organization
- **`commit-commands@claude-plugins-official`** - Streamlines git workflow (commit, push, PR commands). Lightweight convenience.
- **`claude-code-setup@claude-plugins-official`** - Analyzes codebase and recommends hooks, skills, MCP servers, and subagents. Good for initial repo optimization.
- **`hookify@claude-plugins-official`** - Create hooks to prevent unwanted behaviors by analyzing conversation patterns.

### Worth Considering (from third-party marketplaces)
- **`github@claude-plugins-official`** - Official GitHub MCP server for issues, PRs, repo management. Useful for heavy GitHub interaction.
- **`security-guidance@claude-plugins-official`** - Security reminder hook when editing files. Lightweight, low friction.

## Coverage Summary

Current 8 plugins already handle Go, semantic analysis, docs, frontend, code review, and browser debugging. The biggest gaps are **MCP development** and **git workflow** tooling.

## Top 3 Recommended Additions

1. **`mcp-server-dev`** - directly relevant to MCP server work
2. **`claude-code-setup`** - one-time analysis to optimize repo's Claude Code config
3. **`github`** - if using GitHub for issues/PRs

## Available Marketplaces

| Marketplace | Plugins/Skills |
|---|---|
| **anthropics/claude-plugins-official** | 49 total (32 internal + 17 external) |
| **pleaseai/claude-code-plugins** | ~42 total (10 external + ~32 built-in) |
| **secondsky/claude-skills** | 169 skills + 7 commands = 176 total |

## Reducing Manual Approvals

- Add blanket permissions in `.claude/settings.local.json` for trusted plugin tools
- Use `acceptEdits` default mode (already configured for this project)
- Add specific `allow` rules, e.g. `"Bash(claude plugins:*)"`, `"WebFetch(domain:raw.githubusercontent.com)"`
