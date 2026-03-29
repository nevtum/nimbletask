# NimbleTask Todo CLI

A powerful command-line todo management application built by AI, specifically designed for AI agents.

The genesis of this project was to solve memory challenges faced by AI agents when coordinating long, complex tasks between multiple agent sessions. The design prioritizes:

- **Clean CLI interfaces** for terminal-based workflows
- **Token friendly interface** keeping AI agent context windows low
- **Hierarchical organization** for complex projects
- **Flexible configuration** for different use cases
- **Robust error handling** for predictable behavior

## Features

- **Hierarchical Todo Management**: Create parent-child relationships with automatic numbering (1, 1.1, 1.1.1, etc.)
- **Priority Support**: Built-in priority system with configurable defaults
- **Flexible File Locations**: Custom config and todo file paths
- **Shell Completion**: Comprehensive autocompletion for bash, zsh, fish, and PowerShell
- **Clean CLI Interface**: Intuitive command structure with helpful error messages
- **Markdown Output**: Saves todos in readable Markdown format
- **Global Configuration**: Centralized settings management

## Installation

This application requires Go installed on your machine and that $GOPATH value is in your PATH environment variable.

```bash
go install github.com/nevtum/nimbletask@latest
```

## Quick Start

1. **Initialize the configuration** (first time only):
   ```bash
   todo init
   ```

2. **Add your first todo**:
   ```bash
   todo add "Build something amazing"
   ```

3. **List all todos**:
   ```bash
   todo list
   ```

4. **Complete a todo**:
   ```bash
   todo complete <todo-id>
   ```

## Commands

### Global Flags

All commands support these global flags:

- `--config string`: Configuration directory root (default: `~/.config/nimbletask`)
- `--file string`: Path to todo list file (default: `todo.md` in current directory)
- `-h, --help`: Show help for the command

### `add` - Add a new todo item

```bash
todo add [title] [flags]
```

**Flags:**
- `--parent string`: Parent todo ID for hierarchical structure
- `--priority int`: Priority level for the new todo (overrides config default)

**Examples:**
```bash
# Simple todo
todo add "Review pull requests"

# High priority todo
todo add "Fix critical bug" --priority 5

# Subtask under parent
todo add "Update documentation" --parent <id>

# Nested subtask
todo add "Add examples" --parent <id>
```

### `complete` - Mark a todo item as completed

```bash
todo complete [id]
```

**Examples:**
```bash
# Complete a specific todo
todo complete Kl_u9OqgV73aHsKmoYa57

# Complete subtasks propagate to parent
todo complete 1.1.2
```

### `list` - List all todos

```bash
todo list
```

Shows todos in hierarchical format with completion status:
```
1. [ ] <id:abc123> Main project
  1.1 [x] <id:def456> Completed subtask
  1.2 [ ] <id:ghi789> Pending subtask
2. [ ] <id:jkl012> Another task
```

### `init` - Initialize the global configuration file

```bash
todo init [flags]
```

**Examples:**
```bash
# Initialize default config
todo init

# Initialize custom config location
todo init --config /path/to/custom/config
```

### `help` - Get help for any command

```bash
todo help [command]
```

**Examples:**
```bash
# Get general help
todo help

# Get help for specific command
todo help add
todo help complete
```

### `completion` - Generate shell autocompletion

```bash
todo completion [shell]
```

**Supported Shells:**
- `bash`
- `zsh`
- `fish`
- `powershell`

**Examples:**
```bash
# Generate zsh completion
todo completion zsh > ~/.zsh/completion/_todo

# Generate bash completion
todo completion bash > ~/.bash_completion.d/todo
```

## Configuration

### Default Configuration

The application creates a default configuration at `~/.config/nimbletask/config.json`:

```json
{
  "default_priority": 3,
  "filename": "todos.md",
}
```

### Custom Configuration

You can specify custom config locations:

```bash
# Use custom config directory
todo --config /path/to/config add "New todo"

# Use custom todo file
todo --file my-tasks.md add "Another todo"
```

## Usage Patterns

### Project Management

```bash
# Main project
todo add "Build Orders API"

# Features
todo add "User authentication" --parent <project-id>
todo add "Data processing" --parent <project-id>
todo add "API endpoints" --parent <project-id>

# Sub-features
todo add "JWT implementation" --parent <auth-id>
todo add "OAuth support" --parent <auth-id>
```

### Personal Task Management

```bash
todo add "Review research papers" --priority 4
todo add "Update portfolio" --priority 3
todo add "Learn new framework" --priority 5
```

### Code Review Workflow

```bash
todo add "Review PR #123" --priority 4
todo add "Check test coverage" --parent <pr-id>
todo add "Verify documentation" --parent <pr-id>
```

## Advanced Features

### Hierarchical Organization

The application supports complex nested structures:

```
1. [ ] <id:main> AI Project
  1.1 [ ] <id:ml> Machine Learning
    1.1.1 [ ] <id:data-prep>Data Preparation
    1.1.2 [ ] <id:model-training>Model Training
  1.2 [ ] <id:frontend> Frontend
    1.2.1 [ ] <id:ui> UI Components
    1.2.2 [ ] <id:api> API Integration
  1.3 [ ] <id:backend> Backend
    1.3.1 [ ] <id:auth> Authentication
    1.3.2 [ ] <id:db> Database
```

### Priority System

Priorities help organize tasks by importance:

- **Priority 1**: Critical (immediate attention required)
- **Priority 2**: High (important, can wait)
- **Priority 3**: Medium (default priority)
- **Priority 4**: Low (nice to have)
- **Priority 5**: Lowest (when you have spare time)

### Multi-File Management

Different projects can use separate todo files:

```bash
# Work todos
todo --file work.md add "Quarterly report"

# Personal todos
todo --file personal.md add "Plan vacation"

# Open source todos
todo --file oss.md add "Contribute to project"
```

## Development Notes

### Built by AI, for AI

Majority of this tool was generated by AI to solve common memory challenges faced by AI agents. The design prioritizes:

- **Clean CLI interfaces** for terminal-based workflows
- **Token friendly interface** keeping AI agent context windows low
- **Hierarchical organization** for complex projects
- **Flexible configuration** for different use cases
- **Robust error handling** for predictable behavior
