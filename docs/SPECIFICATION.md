# Nested Todo CLI - Implementation Specification

## Global Configuration

## Overview
Build a Go CLI tool for managing hierarchical todo lists with native markdown persistence.

## Data Structure

### Global Configuration File
- The application supports a global configuration file located at `~/.todo/config.json`.
- The configuration file contains:
  - `todo_list_path`: Relative path to the markdown file for the todo list.
  - `default_priority`: Default priority level for new todo items.

### Initialization Command
- A command `todo init-config` initializes the global configuration file.
- This command creates the hidden directory `.todo` and generates a default `config.json`.

### Loading Configuration
- The application loads the global configuration at startup to determine the todo list path, relative to the directory where the command is executed, and default settings.
- If the global configuration file does not exist, the user is prompted to initialize it.

### Cross-Platform Compatibility
- The implementation ensures that the global configuration works across Windows, macOS, and Linux.
- Platform-independent methods are used for home directory retrieval and file path construction.

### Command Modifications
- All commands that modify the todo list reference the global configuration file for the path to create/update the todo list.
- Default settings for new todos are retrieved from the global configuration.

### Error Handling
- The application handles errors gracefully when loading the configuration file or if the specified todo list path is invalid.
- Appropriate error messages are displayed to the user.

### Core Types
```go
type Todo struct {
    ID          string    
    Title       string
    Description string
    Completed   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ParentID    string    
    Children    []*Todo   
    Priority    int       
    DueDate     *time.Time
    Tags        []string
}

type TodoList struct {
    todos    map[string]*Todo
    roots    []*Todo
    modified bool
}
```

### Required Operations
```go
// CRUD
func (tl *TodoList) Add(title string, parentID string, position int) (*Todo, error)
func (tl *TodoList) Get(id string) (*Todo, error)
func (tl *TodoList) Update(id string, updates TodoUpdate) (*Todo, error)
func (tl *TodoList) Delete(id string) error

// Hierarchy
func (tl *TodoList) Move(id string, newParentID string, position int) error
func (tl *TodoList) Promote(id string) error
func (tl *TodoList) Demote(id string, siblingID string) error
func (tl *TodoList) Reorder(id string, newPosition int) error

// Queries
func (tl *TodoList) GetRoots() []*Todo
func (tl *TodoList) GetChildren(parentID string) []*Todo
func (tl *TodoList) GetPath(id string) ([]*Todo, error)
func (tl *TodoList) GetDepth(id string) (int, error)

// Status
func (tl *TodoList) Complete(id string) error
func (tl *TodoList) Uncomplete(id string) error
func (tl *TodoList) Toggle(id string) error
func (tl *TodoList) CompleteSubtree(id string) error

// Validation
func (tl *TodoList) Validate() error
func (tl *TodoList) CanMove(id, newParentID string) (bool, error)

// Persistence
func (tl *TodoList) Save(file *File) error
func (tl *TodoList) Load(file *File) error
```

## Persistence Format

**Extended Task Lists (Option 1)**

```markdown
- [ ] <!-- id:V1StGXR8_Z5jd|parent:|priority:1|created:2024-01-15 --> Project Proposal
  Description text here
  
  - [x] <!-- id:wH9mK2pL4nQ7|parent:V1StGXR8_Z5jd|completed:2024-01-16 --> Research
    - [x] <!-- id:bX5vC3jN8kM1|parent:wH9mK2pL4nQ7|completed:2024-01-16 --> Review sources
  
  - [ ] <!-- id:yZ4wE6hJ9lP2|parent:V1StGXR8_Z5jd|due:2024-02-01 --> Draft
```

**Re-hydration Strategy:**
1. Parse markdown AST
2. Extract todos with metadata from HTML comments
3. Build map[string]*Todo for O(1) lookup
4. Wire up parent-child relationships via ParentID
5. Calculate depth/ancestors as needed

## CLI Design

**Git-style Subcommands**

```bash
# Core operations
todo add "Title" [--parent ID] [--position N] [--priority P] [--due DATE]
todo move ID --parent NEW_PARENT [--position N]
todo complete ID [--recursive]
todo delete ID [--force]
todo edit ID [--title "New"] [--desc "Desc"] [--priority N] [--due DATE]

# Display
todo list [--tree] [--format json|text]
todo get ID [--format markdown|json]
todo search KEYWORD [--status incomplete|completed]

# Help
todo help [COMMAND]
todo help --examples

# All commands support --format json for programmatic access
```

## Display Formats

**1. Numbered Path (Default - Human/AI readable)**
```bash
$ todo list

1. [ ] V1StGXR8_Z5jd Project Proposal
   1.1. [x] wH9mK2pL4nQ7 Research
        1.1.1. [x] bX5vC3jN8kM1 Review sources
   1.2. [ ] yZ4wE6hJ9lP2 Draft
```

**2. JSON Tree (--format json)**
```json
{
  "roots": [
    {
      "id": "V1StGXR8_Z5jd",
      "title": "Project Proposal",
      "path": "1",
      "status": "incomplete",
      "available_operations": ["add-child", "complete", "move", "delete"],
      "children": [...]
    }
  ]
}
```

## Key Design Decisions

1. **ID-based references**: NanoID strings (12 characters, URL-safe base64), not pointers, for serialization safety. 12-character NanoIDs provide ~1% collision probability after 400K IDs—perfectly safe for personal use with collision checking. Much shorter than ULID/UUID for comfortable CLI typing
2. **Position parameter**: 0=first, -1=append, explicit position numbers otherwise
3. **ParentID in storage**: Empty string = root-level
4. **HTML comment metadata**: Hidden, won't affect markdown rendering
5. **Error returns**: All mutating operations validate (cycle detection, invalid IDs)
6. **Two display modes**: Human-friendly tree + machine-friendly JSON
7. **Git-style commands**: Familiar, self-documenting, extensible

## Implementation Notes

- Use standard library where possible
- Support `--format json` for all output commands
- Validate cycles on every move operation
- Maintain order in Children slice
- Update `modified` flag on mutations
- Support relative dates ("today", "tomorrow", "+7d")
