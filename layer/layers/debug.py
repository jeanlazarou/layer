"""
Debug layer implementation for the Layer language.

Provides debugging utilities, logging, and introspection.
"""

import sys
from typing import Any
from datetime import datetime


class DebugLogger:
    """Logger for debug layer."""

    def __init__(self, enabled=True):
        self.enabled = enabled
        self.logs = []

    def log(self, message, level="INFO"):
        """Log a message."""
        if self.enabled:
            timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
            log_entry = f"[{timestamp}] [{level}] {message}"
            print(log_entry, file=sys.stderr)
            self.logs.append(log_entry)

    def get_logs(self):
        """Get all logged messages."""
        return self.logs

    def clear(self):
        """Clear all logs."""
        self.logs = []


# Global debug logger
_debug_logger = DebugLogger()


def setup_debug_layer(layer_context):
    """Set up the debug layer with debugging utilities."""

    def log(message):
        """Log a message."""
        _debug_logger.log(str(message))
        return message

    def trace(name, fn):
        """
        Wrap a function with tracing.

        Logs when the function is called and what it returns.
        """
        def traced(*args):
            args_str = ", ".join(str(arg) for arg in args)
            _debug_logger.log(f"{name}({args_str}) called", "TRACE")
            result = fn(*args)
            _debug_logger.log(f"{name}({args_str}) = {result}", "TRACE")
            return result
        return traced

    def inspect(value):
        """
        Inspect a value and return detailed information.
        """
        info = {
            'type': type(value).__name__,
            'value': value,
        }

        if isinstance(value, (list, tuple)):
            info['length'] = len(value)
            info['elements'] = value

        if isinstance(value, dict):
            info['keys'] = list(value.keys())
            info['size'] = len(value)

        if callable(value):
            info['callable'] = True
            if hasattr(value, '__name__'):
                info['name'] = value.__name__

        _debug_logger.log(f"Inspect: {info}", "DEBUG")
        return value

    def assert_equal(expected, actual, message="Assertion failed"):
        """Assert that two values are equal."""
        if expected != actual:
            error_msg = f"{message}: expected {expected}, got {actual}"
            _debug_logger.log(error_msg, "ERROR")
            raise AssertionError(error_msg)
        _debug_logger.log(f"Assertion passed: {expected} == {actual}", "DEBUG")
        return actual

    def benchmark(name, fn):
        """
        Benchmark a function execution time.
        """
        import time

        def benchmarked(*args):
            start = time.perf_counter()
            result = fn(*args)
            end = time.perf_counter()
            elapsed = (end - start) * 1000  # milliseconds
            _debug_logger.log(f"{name} took {elapsed:.3f}ms", "BENCH")
            return result

        return benchmarked

    def breakpoint_fn(message="Breakpoint"):
        """
        Simulate a breakpoint (logs and pauses for inspection).
        """
        _debug_logger.log(f"BREAKPOINT: {message}", "BREAK")
        # In a real implementation, this could pause execution
        return None

    def get_logs():
        """Get all debug logs."""
        return _debug_logger.get_logs()

    def clear_logs():
        """Clear all debug logs."""
        _debug_logger.clear()
        return None

    # Register debug functions
    layer_context.set_current_layer('debug')

    debug_functions = {
        'log': log,
        'trace': trace,
        'inspect': inspect,
        'assert_equal': assert_equal,
        'benchmark': benchmark,
        'breakpoint': breakpoint_fn,
        'get_logs': get_logs,
        'clear_logs': clear_logs,
    }

    for name, func in debug_functions.items():
        layer_context.define(name, func, 'debug')

    return layer_context
