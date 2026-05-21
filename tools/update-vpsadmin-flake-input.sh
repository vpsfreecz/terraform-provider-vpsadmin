#!/usr/bin/env bash
set -euo pipefail

input="vpsadmin"

usage() {
  cat <<'EOF'
Usage: tools/update-vpsadmin-flake-input.sh [--input <name>]

Update a flake input, verify that only flake.lock changed, and commit the
result using the repository's flake update commit message format.

Defaults to updating the vpsadmin input.
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --input)
      if [ "$#" -lt 2 ]; then
        echo "Missing value for --input" >&2
        exit 2
      fi
      input="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

for cmd in git jq nix; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Missing required command: $cmd" >&2
    exit 1
  fi
done

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "Tracked files must be clean before updating flake input ${input}" >&2
  exit 1
fi

old_rev="$(
  nix flake metadata --json --no-write-lock-file . \
    | jq -er --arg input "$input" '.locks.nodes[$input].locked.rev'
)"

nix flake update "$input"

if git diff --quiet -- flake.lock; then
  echo "flake input ${input} is already up to date at ${old_rev:0:9}"
  exit 0
fi

changed_files="$(git diff --name-only)"
if [ "$changed_files" != "flake.lock" ]; then
  echo "Unexpected tracked changes after updating ${input}:" >&2
  printf '%s\n' "$changed_files" >&2
  exit 1
fi

new_rev="$(
  nix flake metadata --json --no-write-lock-file . \
    | jq -er --arg input "$input" '.locks.nodes[$input].locked.rev'
)"

git add flake.lock

msgfile="$(mktemp)"
trap 'rm -f "$msgfile"' EXIT

cat > "$msgfile" <<EOF
flake: ${input} ${old_rev:0:9} -> ${new_rev:0:9}

Update the ${input} flake input to pick up upstream changes used by the
provider development shell, builds, and integration tests.
EOF

awk '
  length($0) > 80 {
    print FNR ":" length($0) ":" $0
    bad = 1
  }
  END { exit bad }
' "$msgfile"

git commit -F "$msgfile"
