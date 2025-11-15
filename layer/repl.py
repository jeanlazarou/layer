"""
Layer Language REPL

Read-Eval-Print Loop for interactive Layer programming.
"""

import sys
from layer.lexer import tokenize
from layer.parser import parse
from layer.interpreter import Interpreter
from layer import __version__


class LayerREPL:
    """Interactive REPL for Layer language."""

    def __init__(self):
        self.interpreter = Interpreter()
        self.history = []

    def eval_line(self, line: str):
        """Evaluate a single line of Layer code."""
        try:
            tokens = tokenize(line)
            ast = parse(tokens)

            # Execute the program
            result = self.interpreter.execute(ast)

            if result is not None:
                print(result)

        except SyntaxError as e:
            print(f"Syntax Error: {e}", file=sys.stderr)
        except RuntimeError as e:
            print(f"Runtime Error: {e}", file=sys.stderr)
        except Exception as e:
            print(f"Error: {e}", file=sys.stderr)

    def run(self):
        """Run the REPL."""
        print(f"Layer Programming Language v{__version__}")
        print("Type 'exit' or Ctrl+D to quit")
        print("Type 'help' for help")
        print()

        while True:
            try:
                line = input("layer> ")

                if not line.strip():
                    continue

                if line.strip() == 'exit':
                    break

                if line.strip() == 'help':
                    self.print_help()
                    continue

                if line.strip() == 'layers':
                    self.print_layers()
                    continue

                if line.strip() == 'clear':
                    self.interpreter = Interpreter()
                    print("Environment cleared")
                    continue

                self.history.append(line)
                self.eval_line(line)

            except EOFError:
                print("\nGoodbye!")
                break
            except KeyboardInterrupt:
                print("\nKeyboardInterrupt")
                continue

    def print_help(self):
        """Print help information."""
        print("""
Layer REPL Commands:
  help     - Show this help message
  layers   - Show enabled layers
  clear    - Clear the environment
  exit     - Exit the REPL

Layer Syntax:
  fn name(x) = x * 2          # Define function
  let x = 5 in x * 2          # Let binding
  match x with | 0 -> ... # Pattern matching
  [1, 2, 3]                   # List literal
  {key: value}                # Record literal

Layer Directives:
  @layers [core, ui, debug]   # Enable layers
  @layer core { ... }         # Define layer code
        """)

    def print_layers(self):
        """Print information about enabled layers."""
        print(f"Enabled layers: {self.interpreter.layer_context.layer_order}")
        print("\nAvailable layers:")
        print("  core  - Functional programming primitives (always enabled)")
        print("  ui    - User interface components")
        print("  debug - Debugging and logging utilities")


def main():
    """Main REPL entry point."""
    repl = LayerREPL()
    repl.run()


if __name__ == "__main__":
    main()
