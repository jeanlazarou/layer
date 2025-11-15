"""
Layer Language CLI

Command-line interface for running Layer programs.
"""

import sys
import argparse
from pathlib import Path
from layer.interpreter import interpret


def run_file(filepath: str):
    """Run a Layer source file."""
    path = Path(filepath)

    if not path.exists():
        print(f"Error: File not found: {filepath}", file=sys.stderr)
        sys.exit(1)

    if not path.suffix == '.layer':
        print(f"Warning: File does not have .layer extension", file=sys.stderr)

    try:
        source = path.read_text()
        result = interpret(source)

        if result is not None:
            print(result)

    except SyntaxError as e:
        print(f"Syntax Error: {e}", file=sys.stderr)
        sys.exit(1)
    except RuntimeError as e:
        print(f"Runtime Error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    """Main CLI entry point."""
    parser = argparse.ArgumentParser(
        description="Layer Programming Language - A functional language with composable layers"
    )

    subparsers = parser.add_subparsers(dest='command', help='Commands')

    # Run command
    run_parser = subparsers.add_parser('run', help='Run a Layer program')
    run_parser.add_argument('file', help='Layer source file to run')

    # Version command
    subparsers.add_parser('version', help='Show version information')

    args = parser.parse_args()

    if args.command == 'run':
        run_file(args.file)
    elif args.command == 'version':
        from layer import __version__
        print(f"Layer Programming Language v{__version__}")
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
