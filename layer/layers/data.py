"""
Data layer implementation for the Layer language.

Provides file I/O, JSON operations, and localStorage functionality.
Uses Option pattern (Some/None) for error handling.
"""

import os
import json
from typing import Any, Dict


# Simple Option type implementation
class Some:
    """Represents a successful value."""
    def __init__(self, value):
        self.value = value

    def __repr__(self):
        return f"Some({self.value})"

    def __eq__(self, other):
        return isinstance(other, Some) and self.value == other.value


class NoneOption:
    """Represents absence of a value or an error."""
    def __repr__(self):
        return "None"

    def __eq__(self, other):
        return isinstance(other, NoneOption)


# Singleton None instance
_none_instance = NoneOption()


# In-memory localStorage (for non-browser environments)
_local_storage: Dict[str, Any] = {}


def setup_data_layer(layer_context):
    """Set up the data layer with file, JSON, and localStorage operations."""

    # ========== File Operations ==========

    def read_file(path):
        """
        Read a file and return its contents.

        Returns: Some(content) if successful, None if error
        """
        try:
            # Handle relative and absolute paths
            file_path = os.path.abspath(os.path.expanduser(str(path)))

            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            return Some(content)
        except Exception as e:
            # Could log error for debugging
            return _none_instance

    def write_file(path, content):
        """
        Write content to a file.

        Returns: Some(content) if successful, None if error
        """
        try:
            file_path = os.path.abspath(os.path.expanduser(str(path)))

            # Ensure directory exists (only if there's a directory part)
            dir_path = os.path.dirname(file_path)
            if dir_path:
                os.makedirs(dir_path, exist_ok=True)

            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(str(content))

            return Some(content)
        except Exception as e:
            return _none_instance

    def exists(path):
        """
        Check if a file or directory exists.

        Returns: boolean
        """
        try:
            file_path = os.path.abspath(os.path.expanduser(str(path)))
            return os.path.exists(file_path)
        except Exception:
            return False

    def delete_file(path):
        """
        Delete a file.

        Returns: Some(path) if successful, None if error
        """
        try:
            file_path = os.path.abspath(os.path.expanduser(str(path)))

            if os.path.isfile(file_path):
                os.remove(file_path)
                return Some(path)
            else:
                return _none_instance
        except Exception:
            return _none_instance

    def list_files(directory):
        """
        List files in a directory.

        Returns: Some([paths]) if successful, None if error
        """
        try:
            dir_path = os.path.abspath(os.path.expanduser(str(directory)))

            if not os.path.isdir(dir_path):
                return _none_instance

            files = os.listdir(dir_path)
            return Some(files)
        except Exception:
            return _none_instance

    # ========== JSON Operations ==========

    def read_json(path):
        """
        Read and parse a JSON file.

        Returns: Some(parsed_data) if successful, None if error
        """
        try:
            file_path = os.path.abspath(os.path.expanduser(str(path)))

            with open(file_path, 'r', encoding='utf-8') as f:
                data = json.load(f)

            return Some(data)
        except Exception:
            return _none_instance

    def write_json(path, data):
        """
        Write data to a JSON file.

        Returns: Some(data) if successful, None if error
        """
        try:
            file_path = os.path.abspath(os.path.expanduser(str(path)))

            # Ensure directory exists (only if there's a directory part)
            dir_path = os.path.dirname(file_path)
            if dir_path:
                os.makedirs(dir_path, exist_ok=True)

            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2)

            return Some(data)
        except Exception:
            return _none_instance

    def parse_json(json_string):
        """
        Parse a JSON string.

        Returns: Some(parsed_data) if successful, None if error
        """
        try:
            data = json.loads(str(json_string))
            return Some(data)
        except Exception:
            return _none_instance

    def to_json(data):
        """
        Convert data to JSON string.

        Returns: Some(json_string) if successful, None if error
        """
        try:
            json_str = json.dumps(data, indent=2)
            return Some(json_str)
        except Exception:
            return _none_instance

    # ========== localStorage Operations ==========

    def get(key):
        """
        Get a value from localStorage.

        Returns: Some(value) if key exists, None if not found
        """
        key_str = str(key)
        if key_str in _local_storage:
            return Some(_local_storage[key_str])
        else:
            return _none_instance

    def set_key(key, value):
        """
        Set a value in localStorage.

        Returns: Some(value) always (localStorage.set doesn't fail)
        """
        _local_storage[str(key)] = value
        return Some(value)

    def delete_key(key):
        """
        Delete a key from localStorage.

        Returns: Some(key) if existed, None if didn't exist
        """
        key_str = str(key)
        if key_str in _local_storage:
            del _local_storage[key_str]
            return Some(key)
        else:
            return _none_instance

    def keys():
        """
        Get all keys in localStorage.

        Returns: list of keys
        """
        return list(_local_storage.keys())

    def clear_storage():
        """
        Clear all localStorage.

        Returns: None (unit)
        """
        _local_storage.clear()
        return _none_instance

    # ========== Option Type Constructors ==========

    def some(value):
        """Create a Some value."""
        return Some(value)

    def none():
        """Create a None value."""
        return _none_instance

    def is_some(option):
        """Check if option is Some."""
        return isinstance(option, Some)

    def is_none(option):
        """Check if option is None."""
        return isinstance(option, NoneOption)

    def unwrap(option, default=None):
        """
        Unwrap an Option value.

        Returns the value if Some, or default if None.
        """
        if isinstance(option, Some):
            return option.value
        else:
            return default

    # ========== Pure Helper Functions ==========

    def map_option(fn, option):
        """
        Map a function over an Option value.

        If option is Some(value), returns Some(fn(value))
        If option is None, returns None
        """
        if isinstance(option, Some):
            try:
                result = fn(option.value)
                return Some(result)
            except Exception:
                return _none_instance
        else:
            return _none_instance

    def flatmap_option(fn, option):
        """
        FlatMap a function over an Option value.

        If option is Some(value), returns fn(value) (which should return an Option)
        If option is None, returns None
        """
        if isinstance(option, Some):
            try:
                return fn(option.value)
            except Exception:
                return _none_instance
        else:
            return _none_instance

    # Register all functions in the data layer
    layer_context.set_current_layer('data')

    data_functions = {
        # File operations
        'read_file': read_file,
        'write_file': write_file,
        'exists': exists,
        'delete': delete_file,
        'list_files': list_files,

        # JSON operations
        'read_json': read_json,
        'write_json': write_json,
        'parse_json': parse_json,
        'to_json': to_json,

        # localStorage operations
        'get': get,
        'set': set_key,
        'delete_key': delete_key,
        'keys': keys,
        'clear': clear_storage,

        # Option type
        'Some': some,
        'None': none,
        'is_some': is_some,
        'is_none': is_none,
        'unwrap': unwrap,
        'map_option': map_option,
        'flatmap_option': flatmap_option,
    }

    for name, func in data_functions.items():
        layer_context.define(name, func, 'data')

    return layer_context
