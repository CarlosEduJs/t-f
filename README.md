# t-f

A utility to convert CSS variables into Design Tokens (DTCG format).

t-f is a command-line tool designed to bridge the gap between CSS stylesheets and design token systems. It parses CSS custom properties, handles theme variations (like light and dark modes), and generates a structured JSON file following the Design Tokens Community Group (DTCG) specification.

## Features

- Parses CSS custom properties from standard rulesets and Tailwind-style @theme blocks.
- Automatically categorizes tokens based on naming conventions (color, spacing, typography, etc.).
- Resolves variable aliases and basic calc() expressions.
- Converts OKLCH colors to HEX (optional, useful for tools like Figma).
- Handles theme inheritance and overrides.

## Installation

To build the tool from source, ensure you have Go installed and run:

```bash
make build
```

The binary will be created in the .bin directory.

## Usage

Run the tool by providing an input CSS file and the desired output JSON path:

```bash
./.bin/t-f input.css output.json
```

### Figma Compatibility

If you need the output to be compatible with Figma's color requirements, use the --figma flag. This will flatten the token structure and ensure colors are converted to HEX values with specific metadata:

```bash
./.bin/t-f --figma input.css output.json
```

## How it works

1. Parsing: The tool reads your CSS and identifies all custom properties defined in :root, .dark, or @theme blocks.
2. Resolution: It resolves var() references and evaluates calc() expressions to get the final values.
3. Categorization: Based on the variable names (e.g., --color-primary), it groups tokens into categories like colors, spacing, or typography.
4. Export: It generates a JSON file that follows the DTCG standard, making it easy to consume in other design or development tools.

## Development

The project includes a Makefile for common tasks:

- make build: Compiles the project.
- make test: Runs the test suite.
- make lint: Runs code quality checks.
- make install-hooks: Sets up git pre-commit hooks for linting and testing.
