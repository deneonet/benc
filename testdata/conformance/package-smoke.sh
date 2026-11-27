#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
repo="$PWD"
scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT

case "${1:?python or javascript}" in
  python)
    python3 -m venv "$scratch/venv"
    "$scratch/venv/bin/python" -m pip install "$repo/std/python"
    (cd "$scratch" && "$scratch/venv/bin/python" -c '
import bstd
data = bytearray(bstd.size_string("A\0B"))
assert bstd.marshal_string(0, data, "A\0B") == len(data)
assert bstd.unmarshal_string(0, data) == (len(data), "A\0B")
print("Installed Python package: OK")')
    ;;
  javascript)
    npm pack "$repo/std/javascript" --pack-destination "$scratch" --silent
    mkdir -p "$scratch/app"
    npm install --prefix "$scratch/app" "$scratch"/*.tgz --ignore-scripts --silent
    (cd "$scratch/app" && node -e '
const b = require("@banditmoscow1337/benc");
const data = new Uint8Array(b.sizeString("A\0B"));
if (b.marshalString(0, data, "A\0B") !== data.length ||
    b.unmarshalString(0, data)[1] !== "A\0B") throw Error("installed runtime mismatch");
console.log("Installed JS/TS package: OK");')
    cat > "$scratch/app/consumer.ts" <<'EOF'
import { Time, sizeUint } from '@banditmoscow1337/benc';
const instant: Time = new Time(1n);
const size: number = sizeUint(instant.unixNano);
if (size < 1) throw Error('invalid size');
EOF
    (cd "$scratch/app" && npx --yes --package typescript@5.9.3 tsc \
      --strict --noEmit --target es2020 --module nodenext --moduleResolution nodenext consumer.ts)
    ;;
  *) echo "unknown package: $1" >&2; exit 2 ;;
esac
