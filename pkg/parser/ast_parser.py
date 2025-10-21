#!/usr/bin/env python3
"""
Python AST Parser for Infrar Engine

This script parses Python source code and outputs JSON representation
of the AST for consumption by the Go-based transformation engine.
"""

import ast
import json
import sys
from typing import Any, Dict, List, Optional


def get_value_type(value: Any) -> str:
    """
    Map Python types to Infrar value types.

    Python type -> Infrar type:
    - str -> string
    - int, float -> number
    - bool -> bool
    - None -> none
    """
    if value is None:
        return "none"
    elif isinstance(value, bool):  # Must check bool before int (bool is subclass of int)
        return "bool"
    elif isinstance(value, str):
        return "string"
    elif isinstance(value, (int, float)):
        return "number"
    else:
        return "unknown"


def extract_value(node: ast.AST) -> Dict[str, Any]:
    """Extract value from an AST node."""
    if isinstance(node, ast.Constant):
        # Python 3.8+
        value_type = get_value_type(node.value)
        return {
            "type": value_type,
            "value": node.value
        }
    elif isinstance(node, ast.Str):
        # Python <3.8
        return {"type": "string", "value": node.s}
    elif isinstance(node, ast.Num):
        # Python <3.8
        return {"type": "number", "value": str(node.n)}
    elif isinstance(node, ast.NameConstant):
        # Python <3.8 (True, False, None)
        return {"type": get_value_type(node.value), "value": node.value}
    elif isinstance(node, ast.Name):
        return {"type": "variable", "value": node.id}
    elif isinstance(node, ast.List):
        return {
            "type": "list",
            "value": [extract_value(elt) for elt in node.elts]
        }
    elif isinstance(node, ast.Dict):
        return {
            "type": "dict",
            "value": {
                extract_value(k)["value"]: extract_value(v)
                for k, v in zip(node.keys, node.values)
            }
        }
    else:
        return {"type": "unknown", "value": None}


def extract_imports(tree: ast.Module) -> List[Dict[str, Any]]:
    """Extract import statements from the AST."""
    imports = []

    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                imports.append({
                    "module": alias.name,
                    "names": [alias.name],
                    "alias": alias.asname or "",
                    "lineno": node.lineno
                })

        elif isinstance(node, ast.ImportFrom):
            module = node.module or ""
            names = [alias.name for alias in node.names]
            imports.append({
                "module": module,
                "names": names,
                "alias": "",
                "lineno": node.lineno
            })

    return imports


def extract_calls(tree: ast.Module, source_lines: List[str]) -> List[Dict[str, Any]]:
    """Extract function calls from the AST, focusing on potential Infrar SDK calls."""
    calls = []

    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            call_info = {
                "lineno": node.lineno,
                "col_offset": node.col_offset,
                "function": None,
                "module": None,
                "arguments": {},
            }

            # Determine the function being called
            if isinstance(node.func, ast.Name):
                # Direct function call: upload(...)
                call_info["function"] = node.func.id

            elif isinstance(node.func, ast.Attribute):
                # Attribute call: infrar.storage.upload(...)
                # or storage.upload(...)
                call_info["function"] = node.func.attr

                # Try to extract the full module path
                parts = []
                current = node.func.value
                while isinstance(current, ast.Attribute):
                    parts.insert(0, current.attr)
                    current = current.value
                if isinstance(current, ast.Name):
                    parts.insert(0, current.id)
                    call_info["module"] = ".".join(parts)

            # Extract arguments
            # Positional arguments
            for i, arg in enumerate(node.args):
                call_info["arguments"][f"arg_{i}"] = extract_value(arg)

            # Keyword arguments
            for keyword in node.keywords:
                call_info["arguments"][keyword.arg] = extract_value(keyword.value)

            # Extract source code snippet
            if 0 <= node.lineno - 1 < len(source_lines):
                call_info["source_code"] = source_lines[node.lineno - 1].strip()

            calls.append(call_info)

    return calls


def parse_python_code(source_code: str) -> Dict[str, Any]:
    """
    Parse Python source code and return JSON representation.

    Args:
        source_code: Python source code as string

    Returns:
        Dictionary with AST information
    """
    try:
        tree = ast.parse(source_code)
        source_lines = source_code.split('\n')

        result = {
            "language": "python",
            "imports": extract_imports(tree),
            "calls": extract_calls(tree, source_lines),
            "source_code": source_code,
            "success": True,
            "error": None
        }

        return result

    except SyntaxError as e:
        return {
            "success": False,
            "error": {
                "type": "SyntaxError",
                "message": str(e),
                "lineno": e.lineno,
                "offset": e.offset,
                "text": e.text
            }
        }
    except Exception as e:
        return {
            "success": False,
            "error": {
                "type": type(e).__name__,
                "message": str(e)
            }
        }


def main():
    """Main entry point - reads from stdin, outputs JSON to stdout."""
    if len(sys.argv) > 1:
        # Read from file if provided
        with open(sys.argv[1], 'r') as f:
            source_code = f.read()
    else:
        # Read from stdin
        source_code = sys.stdin.read()

    result = parse_python_code(source_code)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
