#!/bin/sh
set -eu

repo_path="${POSTS_REPO_PATH:-/app/my_posts_data}"
export HOME="${HOME:-/app}"
export TMPDIR="${TMPDIR:-/tmp}"

if [ "$(id -u)" = "0" ]; then
	mkdir -p "${repo_path}"
	# Docker named volumes start as root:root; poster (uid 1000) needs write access.
	if [ "$(stat -c '%u:%g' "${repo_path}")" != "1000:1000" ]; then
		chown -R poster:poster "${repo_path}"
	fi
	exec su-exec poster:poster env HOME="${HOME}" TMPDIR="${TMPDIR}" "$@"
fi

exec "$@"
