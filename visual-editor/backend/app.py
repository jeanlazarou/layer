"""
Flask backend for Layer Visual Editor

Provides REST API for executing Layer code and capturing output.
"""

from flask import Flask, request, jsonify
from flask_cors import CORS
import sys
from io import StringIO
import traceback

# Add parent directory to path to import layer module
sys.path.insert(0, '../../')

from layer.interpreter import interpret
from layer.file_parser import parse_layer_file, LayerFileParseError

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
def parse_layer_file_endpoint():
    """Parse a .layer file and return structured layer data."""
    try:
        data = request.get_json()
        content = data.get('content', '')

        if not content:
            return jsonify({
                'success': False,
                'error': 'No content provided'
            }), 400

        # Use the centralized parser from layer module
        layer_data = parse_layer_file(content)

        return jsonify({
            'success': True,
            **layer_data.to_dict()
        })

    except LayerFileParseError as e:
        return jsonify({
            'success': False,
            'error': str(e)
        }), 400

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
