#!/usr/bin/env bash
# Determine the next SemVer from commit subjects since the last tag.

# Exit immediately on error, treat unset variables as errors, and fail on pipe errors.
set -euo pipefail

# Semantic versioning rules (case-insensitive):
#   - any commit subject starting with "major:" -> MAJOR bump
#   - else any starting with "feat:"            -> MINOR bump
#   - else any starting with "fix:"             -> PATCH bump
#   - else any other commits                    -> PATCH bump (default)
# Prefixes may be scoped, e.g. feat(api): add endpoint.
# If no tag exists, current version is 1.0.0 (initial release).
# When bumping MINOR, reset PATCH to 0; when bumping MAJOR, reset MINOR/PATCH to 0.
# If there are no new commits since the last tag, do not release.
#
# Standard outputs (stdout by default):
#   LAST_TAG=... (normalized version without v prefix)
#   NEXT_TAG=... (new version with v prefix, e.g., v1.2.3)
#   BUMP=major|minor|patch|none
#   HAS_RELEASE=1|0
#
# If called with --emit-gh-output, writes outputs to $GITHUB_OUTPUT for GitHub Actions.

emit_gh_output=false
if [[ "${1:-}" == "--emit-gh-output" ]]; then
  emit_gh_output=true
fi

found_tag=true
last_tag=""
if last_tag=$(git tag --merged HEAD --sort=-version:refname --list 'v[0-9]*.[0-9]*.[0-9]*' | head -n 1); then
  if [[ -z "$last_tag" ]]; then
    found_tag=false
    last_tag="1.0.0"
  fi
else
  found_tag=false
  last_tag="1.0.0"
fi

normalize_version() {
  echo "$1" | sed 's/^v//'
}

curr_ver=$(normalize_version "$last_tag")

if $found_tag; then
  range="${last_tag}..HEAD"
  mapfile -t subjects < <(git log --format=%s ${range})
else
  mapfile -t subjects < <(git log --first-parent --format=%s HEAD)
fi

if [[ ${#subjects[@]} -eq 0 ]]; then
  bump="none"
else
  bump="patch"

  for s in "${subjects[@]}"; do
    if printf '%s' "$s" | grep -Eiq '^major(:|\([^)]+\):)'; then
      bump="major"
      break
    fi
  done

  if [[ $bump == "patch" ]]; then
    for s in "${subjects[@]}"; do
      if printf '%s' "$s" | grep -Eiq '^feat(:|\([^)]+\):)'; then
        bump="minor"
        break
      fi
    done
  fi

  if [[ $bump == "patch" ]]; then
    for s in "${subjects[@]}"; do
      if printf '%s' "$s" | grep -Eiq '^fix(:|\([^)]+\):)'; then
        bump="patch"
        break
      fi
    done
  fi
fi

IFS='.' read -r major minor patch <<<"$curr_ver"

next_ver="$curr_ver"
case "$bump" in
  major)
    major=$((major+1)); minor=0; patch=0;
    next_ver="$major.$minor.$patch";;
  minor)
    minor=$((minor+1)); patch=0;
    next_ver="$major.$minor.$patch";;
  patch)
    patch=$((patch+1));
    next_ver="$major.$minor.$patch";;
  none)
    :;;
esac

has_release=0
if ! $found_tag; then
  has_release=1
  bump="none"
  next_ver="$curr_ver"
elif [[ "$bump" != "none" ]]; then
  has_release=1
fi

if $emit_gh_output; then
  {
    echo "last_tag=$curr_ver"
    echo "next_tag=v$next_ver"
    echo "bump=$bump"
    echo "has_release=$has_release"
  } >>"$GITHUB_OUTPUT"
else
  printf '%s\n' \
    "LAST_TAG=$curr_ver" \
    "NEXT_TAG=v$next_ver" \
    "BUMP=$bump" \
    "HAS_RELEASE=$has_release"
fi
