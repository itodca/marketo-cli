"""PyInstaller hook for the mrkto CLI.

The CLI lazily imports the HTTP client, config module, and resource modules
to keep startup fast. PyInstaller cannot discover those imports statically, so
we include them here.
"""

from PyInstaller.utils.hooks import collect_submodules

hiddenimports = [
    "mrkto.client",
    "mrkto.config",
    "mrkto.resources",
    *collect_submodules("mrkto.resources"),
]
