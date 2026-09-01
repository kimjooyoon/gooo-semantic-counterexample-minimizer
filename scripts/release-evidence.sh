#!/usr/bin/env bash
set -euo pipefail

bin=${1:?path to the built minimizer binary is required}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
out=${2:?caller-owned release evidence directory is required}
mkdir -p "$out"
for source in "$root"/examples/counterexamples/*.gooo; do
  scenario=$(basename "$source" .gooo)
  mkdir -p "$out/$scenario"
  "$bin" run --meta "$root/.gooo/counterexample-minimizer.gooo" --contract "$root/contracts/denominator-v1.json" --source "$source" --out "$out/$scenario" >/dev/null
done
