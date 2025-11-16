"""
Flask backend for Layer Visual Editor

Provides REST API for executing Layer code and capturing output.
"""

from flask import Flask, request, jsonify
from flask_cors import CORS
import sys
from io import StringIO
import traceback
import re

# Add parent directory to path to import layer module
sys.path.insert(0, '../../')

from layer.interpreter import interpret

app = Flask(__name__)
CORS(app)  # Enable CORS for React frontend


@app.route('/api/execute', methods=['POST'])
def execute_code():
    """Execute Layer code and return result with logs."""
    try:
        data = request.get_json()
        code = data.get('code', '')

        if not code:
            return jsonify({
                'success': False,
                'error': 'No code provided'
            }), 400

        # Capture stderr (for debug logs)
        stderr_capture = StringIO()
        old_stderr = sys.stderr
        sys.stderr = stderr_capture

        try:
            # Execute the Layer code
            result = interpret(code)

            # Get captured logs
            logs = stderr_capture.getvalue()
            log_lines = [line.strip() for line in logs.split('\n') if line.strip()]

            # Convert result to string for JSON serialization
            result_str = str(result) if result is not None else 'None'

            return jsonify({
                'success': True,
                'result': result_str,
                'logs': log_lines
            })

        except SyntaxError as e:
            return jsonify({
                'success': False,
                'error': f'Syntax Error: {str(e)}'
            })

        except RuntimeError as e:
            return jsonify({
                'success': False,
                'error': f'Runtime Error: {str(e)}'
            })

        except Exception as e:
            return jsonify({
                'success': False,
                'error': f'Error: {str(e)}\n{traceback.format_exc()}'
            })

        finally:
            # Restore stderr
            sys.stderr = old_stderr

    except Exception as e:
        return jsonify({
            'success': False,
            'error': f'Server error: {str(e)}'
        }), 500


@app.route('/api/parse', methods=['POST'])
def parse_layer_file():
    """Parse a .layer file and return structured layer data."""
    try:
        data = request.get_json()
        content = data.get('content', '')

        if not content:
            return jsonify({
                'success': False,
                'error': 'No content provided'
            }), 400

        # Parse @layers directive
        layers_match = re.search(r'@layers\s*\[([^\]]+)\]', content)
        if not layers_match:
            return jsonify({
                'success': False,
                'error': 'No @layers directive found'
            }), 400

        enabled_layer_names = [name.strip() for name in layers_match.group(1).split(',')]

        # Parse @layer blocks using brace counting
        parsed_layers = {}
        layer_header_pattern = re.compile(r'@layer\s+(\w+)\s*\{')

        for match in layer_header_pattern.finditer(content):
            layer_name = match.group(1)
            start_pos = match.end()

            # Find matching closing brace by counting
            brace_count = 1
            end_pos = start_pos

            while brace_count > 0 and end_pos < len(content):
                if content[end_pos] == '{':
                    brace_count += 1
                elif content[end_pos] == '}':
                    brace_count -= 1
                end_pos += 1

            if brace_count == 0:
                # Extract layer code (excluding the closing brace)
                layer_code = content[start_pos:end_pos - 1]

                # Remove indentation (typically 4 spaces)
                lines = layer_code.split('\n')
                dedented_lines = []
                for line in lines:
                    if line.startswith('    '):
                        dedented_lines.append(line[4:])
                    else:
                        dedented_lines.append(line)

                parsed_layers[layer_name] = '\n'.join(dedented_lines).strip()

        # Extract main expression (everything after last @layer block)
        last_layer_matches = list(layer_header_pattern.finditer(content))
        main_expression = ''

        if last_layer_matches:
            last_match = last_layer_matches[-1]
            last_layer_start = last_match.start()

            # Find the closing brace of the last layer
            after_last_layer = content[last_layer_start:]
            layer_start_brace = after_last_layer.index('{') + 1
            brace_count = 1
            pos = layer_start_brace

            while brace_count > 0 and pos < len(after_last_layer):
                if after_last_layer[pos] == '{':
                    brace_count += 1
                elif after_last_layer[pos] == '}':
                    brace_count -= 1
                pos += 1

            main_start = last_layer_start + pos
            main_expression = content[main_start:].strip()

            # Remove leading comment lines
            lines = main_expression.split('\n')
            main_lines = []
            for line in lines:
                trimmed = line.strip()
                if trimmed and not trimmed.startswith('#'):
                    main_lines.append(line)

            main_expression = '\n'.join(main_lines).strip()

        return jsonify({
            'success': True,
            'enabledLayers': enabled_layer_names,
            'layers': parsed_layers,
            'mainExpression': main_expression
        })

    except Exception as e:
        return jsonify({
            'success': False,
            'error': f'Parse error: {str(e)}\n{traceback.format_exc()}'
        }), 500


@app.route('/api/health', methods=['GET'])
def health_check():
    """Health check endpoint."""
    return jsonify({
        'status': 'healthy',
        'service': 'Layer Visual Editor Backend'
    })


if __name__ == '__main__':
    print("🚀 Layer Visual Editor Backend Starting...")
    print("📡 API available at http://localhost:5000")
    print("🔗 Frontend should connect to this server")
    app.run(debug=True, port=5000, host='0.0.0.0')
