#!/bin/bash
# Build the maple binary and install it to ~/.local/bin (override with PREFIX).
set -euo pipefail
PREFIX="${PREFIX:-$HOME/.local/bin}"
mkdir -p "$PREFIX"
go build -o "$PREFIX/maple" .
echo "Installed maple to $PREFIX/maple"
echo "Ensure $PREFIX is on your PATH."
