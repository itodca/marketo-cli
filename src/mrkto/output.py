"""Output formatting — JSON, compact, field selection."""

import json
import sys


def print_result(data, fmt="json", fields=None, file=None):
    if file is None:
        file = sys.stdout
    if fields and isinstance(data, list):
        data = [{k: v for k, v in item.items() if k in fields} for item in data]
    elif fields and isinstance(data, dict):
        data = {k: v for k, v in data.items() if k in fields}

    if fmt == "json":
        json.dump(data, file, indent=2, default=str)
        file.write("\n")
    elif fmt == "compact":
        if isinstance(data, list):
            for item in data:
                file.write(json.dumps(item, default=str) + "\n")
        else:
            file.write(json.dumps(data, default=str) + "\n")
    elif fmt == "raw":
        json.dump(data, file, default=str)
        file.write("\n")


def print_error(message, file=None):
    if file is None:
        file = sys.stderr
    print(json.dumps({"error": message}), file=file)
