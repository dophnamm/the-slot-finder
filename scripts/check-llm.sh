#!/bin/sh
# Sends one tiny request to the model provider chosen in server/cmd/server/.env and reports the
# HTTP status, in words, plus the rate-limit headers when the provider sends them.
# The key is passed to curl on stdin, never on the command line, and is never printed.
# Usage: scripts/check-llm.sh   (or: make check-llm)

ENV_FILE="$(cd "$(dirname "$0")/.." && pwd)/server/cmd/server/.env"

# Same defaults as server/internal/constants/ai.go; change them together.
DEFAULT_GEMINI_MODEL="gemini-3.5-flash"
DEFAULT_GROQ_MODEL="openai/gpt-oss-120b"
DEFAULT_OPENROUTER_MODEL="openrouter/free"
GROQ_URL="https://api.groq.com/openai/v1/chat/completions"
OPENROUTER_URL="https://openrouter.ai/api/v1/chat/completions"
GEMINI_URL="https://generativelanguage.googleapis.com/v1beta/models/"

# env_value NAME: the last NAME=value line of the .env file, without quotes or a trailing CR.
env_value() {
	[ -f "$ENV_FILE" ] || return 0
	grep "^$1=" "$ENV_FILE" | tail -n 1 | cut -d= -f2- | tr -d '\r' | sed -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'\$/\1/"
}

provider="$(env_value LLM_PROVIDER)"
provider="${provider:-gemini}"
case "$provider" in
gemini)
	key="$(env_value GEMINI_API_KEY)"
	model="$(env_value GEMINI_MODEL)"
	model="${model:-$DEFAULT_GEMINI_MODEL}"
	url="${GEMINI_URL}${model}:generateContent"
	auth="x-goog-api-key: $key"
	body='{"contents":[{"parts":[{"text":"hi"}]}],"generationConfig":{"maxOutputTokens":5}}'
	;;
groq | openrouter)
	if [ "$provider" = groq ]; then
		key="$(env_value GROQ_API_KEY)"
		model="$(env_value GROQ_MODEL)"
		model="${model:-$DEFAULT_GROQ_MODEL}"
		url="$GROQ_URL"
	else
		key="$(env_value OPENROUTER_API_KEY)"
		model="$(env_value OPENROUTER_MODEL)"
		model="${model:-$DEFAULT_OPENROUTER_MODEL}"
		url="$OPENROUTER_URL"
	fi
	auth="Authorization: Bearer $key"
	body="{\"model\":\"$model\",\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}],\"max_tokens\":5}"
	;;
*)
	echo "Unknown LLM_PROVIDER '$provider' (use gemini, groq or openrouter)."
	exit 1
	;;
esac

echo "Provider: $provider"
echo "Model:    $model"
if [ -z "$key" ]; then
	echo "Not configured: the API key of $provider is empty in server/cmd/server/.env."
	exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
status="$(printf 'header = "%s"\n' "$auth" | curl -s -K - -m 30 -X POST -H 'Content-Type: application/json' \
	-d "$body" -D "$tmp/headers" -o "$tmp/body" -w '%{http_code}' "$url")"

header() { grep -i "^$1:" "$tmp/headers" | tail -n 1 | cut -d: -f2- | tr -d '\r' | sed 's/^ *//'; }

case "$status" in
200) meaning="ok, the key and the model work" ;;
401 | 403) meaning="key rejected" ;;
404) meaning="model unavailable to this key" ;;
429) meaning="rate limited, retry after $(header retry-after)s" ;;
402) meaning="no credits" ;;
503) meaning="provider busy, try again later" ;;
000) meaning="no answer (network error or timeout)" ;;
*) meaning="unexpected status" ;;
esac
echo "Status:   $status ($meaning)"

if [ "$provider" = gemini ]; then
	echo "Limits:   shown in Google AI Studio (aistudio.google.com)."
else
	grep -i '^x-ratelimit-' "$tmp/headers" | tr -d '\r' | sed 's/^/Limits:   /'
fi

if [ "$status" != 200 ]; then
	echo "Detail:   $(head -c 300 "$tmp/body" | tr '\n' ' ')"
	exit 1
fi
