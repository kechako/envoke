# envoke

A CLI tool for efficiently managing environment variables. Manage multiple environment configurations and run commands with environment-specific variable values.

## Features

- Manage multiple environment configurations (development, staging, production, etc.)
- Add, update, delete, and list environment variables
- Import/export .env files
- Hierarchical management with global and environment-specific settings
- Variable expansion functionality
- SQLite-based local database

## Installation

```bash
go install github.com/kechako/envoke/cmd/envoke@latest
```

## Usage

### Basic Commands

```bash
# Create an environment
envoke create development

# Add an environment variable
envoke var add -e development DATABASE_URL "postgres://localhost/myapp_dev"

# List environment variables
envoke var list -e development

# Run a command in the specified environment
envoke run -e development npm start
```

### Environment Management

```bash
# List environments
envoke list

# Create environment
envoke create <environment_name>

# Remove environment
envoke remove <environment_name>

# Rename environment
envoke rename <old_name> <new_name>

# Copy environment
envoke copy <source> <destination>

# Update environment (change description)
envoke update <environment_name>
```

### Variable Management

```bash
# Add variable
envoke var add -e <environment> <name> <value>

# Update variable
envoke var update -e <environment> <name> <new_value>

# Remove variable
envoke var remove -e <environment> <name>

# List variables
envoke var list -e <environment>

# Import from .env file
envoke var import -e <environment> .env

# Export to .env file
envoke var export -e <environment> .env
```

### Command Execution

```bash
# Run command in specified environment
envoke run -e <environment> [--] <command> [args...]

# Example: Start server in development environment
envoke run -e development npm run dev

# Example: Run migration in production environment
envoke run -e production ./migrate up
```

## Configuration

The configuration file is located at `~/.config/envoke/config.yaml`.

```yaml
db_path: /path/to/custom/database.db # Optional
```

The database file is stored at `~/.local/share/envoke/data.db`.

## Global Environment

A special environment called `global` is automatically created, allowing you to set variables common to all environments. Environment-specific variables take precedence, but global variables are also available.

## Variable Expansion

Variables with the `expand` flag can reference other variables:

```bash
envoke var add -e development BASE_URL "https://api.example.com"
envoke var add -e development API_URL '${BASE_URL}/v1' --expand
```

## License

MIT License
