<!-- updated: 2026-02-15T21:30:00Z -->
# Layout - Responsive Layout System

Grid-based responsive layout for terminal dimensions.

## Structure

```
layout/
├── layout.go          # Layout type enum, GetLayout()
├── grid.go            # Grid container
├── grid_row.go        # Row within grid
├── grid_cell.go       # Cell within row
├── region.go          # Named regions for content
└── row_region_params.go # Row configuration parameters
```

## Key Types

| Type | Description |
|------|-------------|
| `Layout` | Enum: Compact, Normal, Wide |
| `Grid` | Container with rows and cells |
| `Region` | Named area for component placement |

## Breakpoints

| Width | Layout | Behavior |
|-------|--------|----------|
| <80 | Compact | Header + services only |
| 80-159 | Normal | Stacked sections |
| ≥160 | Wide | Side-by-side panels |

## Usage

```go
size := terminal.Size{Cols: w, Rows: h}
layout := terminal.GetLayout(size)
// Render based on layout type
```

## Constraints

- Pure calculation, no terminal I/O
- Handles edge cases (0 width, very small)
