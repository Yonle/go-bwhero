#!/usr/bin/env bash
set -euo pipefail

args=()

[[ -n "${LISTEN:-}" ]] && args+=("-listen" "$LISTEN")
[[ -n "${CONCURRENCY_LEVEL:-}" ]] && args+=("-vipsConcurrencyLevel" "$CONCURRENCY_LEVEL")
[[ -n "${IMAGESIZELIMIT:-}" ]] && args+=("-imgSizeLimit" "$IMAGESIZELIMIT")
[[ -n "${ANIMATIONSIZELIMIT:-}" ]] && args+=("-animSizeLimit" "$ANIMATIONSIZELIMIT")
[[ -n "${VIDEOSIZELIMIT:-}" ]] && args+=("-videoSizeLimit" "$VIDEOSIZELIMIT")
[[ -n "${USER_AGENT:-}" ]] && args+=("-userAgent" "$USER_AGENT")
[[ -n "${WORKERS:-}" ]] && args+=("-workers" "$WORKERS") 

exec ${GO_BWHERO_PATH:-`command -v go-bwhero`} "${args[@]}" "$@"
