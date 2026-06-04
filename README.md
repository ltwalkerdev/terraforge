# Terraforge

A terminal UI for managing Terragrunt stacks, built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Install

```bash
cd terraforge
CGO_ENABLED=0 go build -o tforge .

# Make available globally
sudo cp tforge /usr/local/bin/tforge
```

## Usage

```bash
# Defaults to /workspaces/live (no args needed in devcontainer)
tforge

# Or specify a workspace
tforge --workspace /path/to/workspace

# Check version
tforge --version
```

## Layout

```
╭─ Stacks ──────────────────────────────────────────────╮
│  cluster-1/prod  cluster-2/staging  cluster-3/prod    │
╰───────────────────────────────────────────────────────╯
╭─ Units ───────────────────────────────────────────────╮
│  [ ] vpc              ● clean                         │
│  [ ] rke2             ● changed                       │
│  [ ] monitoring       ○                               │
╰───────────────────────────────────────────────────────╯
 /:filter  v:select  .:ellipsis  p:plan  a:apply  d:destroy  S:generate  C:clean  F:output  T:cli  ?:help  [0/15]
╭─ Output ──────────────────────────────────────────────╮
│  waiting for command output...                        │
╰───────────────────────────────────────────────────────╯
```

## Keybindings

### Navigation (vim)

| Key | Action |
|-----|--------|
| `j`/`k` | Up / down |
| `h`/`l` | Back / drill in |
| `gg`/`G` | Top / bottom |
| `ctrl+u`/`ctrl+d` | Half page up / down |
| `H`/`L` | Prev / next stack |
| `/` | Filter units (type to search, `Enter` to keep, `Esc` to clear) |
| `Esc` | Clear active filter |
| `q` | Quit (or back from detail) |
| `?` | Help popup |

### Stack View

| Key | Action |
|-----|--------|
| `v`/`space` | Toggle unit selection |
| `.` | Cycle ellipsis mode (DAG-aware) |
| `A` | Select all |
| `N` | Select none |
| `p` | `terragrunt stack run plan` |
| `a` | `terragrunt stack run apply` |
| `d` | `terragrunt stack run destroy` (confirm) |
| `S` | `terragrunt stack generate` |
| `C` | `terragrunt stack clean` |
| `I` | `terragrunt stack run init` |
| `b` | `terragrunt backend bootstrap --all` |
| `f` | View stack files |
| `F` | Toggle fullscreen output |
| `T` | Toggle CLI command preview |
| `ctrl+x` | Cancel running command |

### Unit Detail (after drill-in)

| Key | Action |
|-----|--------|
| `i` | `terragrunt run -- init` |
| `p` | `terragrunt run -- plan` |
| `a` | `terragrunt run -- apply` |
| `d` | `terragrunt run -- destroy` (confirm) |
| `r` | `apply -refresh-only` |
| `V` | `validate` |
| `e` | `terragrunt render` (resolved config) |
| `s` | `state list` |
| `o` | `output` |
| `l`/`enter` | `state show` (selected resource) |
| `x` | `state rm` (confirm) |
| `R` | `apply -replace=ADDR` |
| `m` | `import ADDR ID` |
| `M` | `state mv SRC DST` |
| `U` | `force-unlock LOCK_ID` |
| `C` | Clean unit .terragrunt-cache |
| `f` | View unit files (HCL, scripts, .tf) |
| `tab`/`shift+tab` | Cycle tabs (State/Outputs/Plan) |

### Ellipsis (DAG-aware selection)

Units start with nothing selected. Use `v`/`space` for manual selection, or `.` for dependency-aware selection:

| Indicator | Meaning | Terragrunt flag |
|-----------|---------|-----------------|
| `[▸]` | Unit + its dependencies (upstream) | `--filter "vpc..."` |
| `[◂]` | Unit + its dependents (downstream) | `--filter "...vpc"` |
| `[x]` | Exact match (resolved dep) | `--filter vpc` |

One ellipsis anchor at a time. Pressing `.` resolves the full transitive dependency graph and selects only the units that will run. Pressing `.` again cycles to dependents, then resets to all selected. Pressing `.` on a different unit replaces the anchor.

Pressing `v`/`space` while in ellipsis mode exits back to manual selection.

### CLI Preview (`T`)

Toggles a line between the status bar and output pane showing the resolved `--filter` flags. Updates to the full command (`terragrunt stack run plan --filter "vpc..."`) after you trigger a run. Helps learn terragrunt's native CLI syntax.

### Fullscreen Output (`F`)

Expands the output pane to the full terminal. Vim-style scrolling (`j`/`k`, `gg`/`G`, `ctrl+u`/`ctrl+d`). Press `F` or `Esc` to return.

## Session Persistence

Selection state and the last 500 lines of output are saved per stack to `~/.local/share/terraforge/`. On restart, your selections and previous output are restored.

Configure the number of persisted log lines in `~/.config/terraforge/config.json`:

```json
{ "logLines": 500 }
```

Set to `0` to disable log persistence.

## File Viewer

Press `f` to browse files associated with the current context:

- **Stack view**: `terragrunt.stack.hcl`
- **Unit view**: `terragrunt.hcl`, `terragrunt.values.hcl`, scripts (`.sh`), and module `.tf` files (after init)

The viewer has basic HCL/TF syntax highlighting and full vim-style navigation.

## How It Works

- Discovers stacks by scanning for `terragrunt.stack.hcl` files
- Discovers units by walking `.terragrunt-stack/` cache directories
- Parses `dependency` and `dependencies` blocks in unit `terragrunt.hcl` for the DAG
- Builds reverse edges (dependents) at startup for downstream resolution
- All commands run from the stack directory via `direnv exec <stack-dir>` ensuring correct env vars
- Unit statuses update after plan/apply (clean, changed, error, running)
- Output is color-coded: white (normal), yellow (warnings), red (errors), green/red/yellow (plan diff)
- Dynamic HCL expressions in dependency paths (ternaries) are displayed as alternatives (e.g. `msk|kafka`)

## Requirements

- Terragrunt v1.0+
- direnv
- Go 1.24+ (build only)
