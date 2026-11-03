#!/bin/bash
# Builds a file-based Go module proxy from the host module cache
#
# Usage: go_modproxy.sh <output dir>
set -euo pipefail

out=${1:?usage: $0 <output dir>}
cache="$(go env GOMODCACHE)/cache/download"

# Make sure every module the build needs is in the host cache
go mod download

rm -rf "$out"
mkdir -p "$out"

# GoMod points at <cache>/<module>/@v/<version>.mod for every module in the graph,
# including replacements; copy its .info, .mod and .zip siblings when present
go list -m -f '{{if .Replace}}{{.Replace.GoMod}}{{else}}{{.GoMod}}{{end}}' all |
	grep "^$cache/" |
	while read -r gomod; do
		base="${gomod%.mod}"
		dest="$out/${base#"$cache"/}"
		mkdir -p "$(dirname "$dest")"
		for ext in info mod zip; do
			if [ -f "$base.$ext" ]; then
				cp "$base.$ext" "$dest.$ext"
			fi
		done
	done

echo "--> Module proxy written to $out ($(du -sh "$out" | cut -f1))"
