#!/usr/bin/env pwsh
# Determine the next SemVer from commit subjects since the last tag.

$ErrorActionPreference = "Stop"

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
# If called with --emit-gh-output, writes outputs to $env:GITHUB_OUTPUT for GitHub Actions.

$emitGhOutput = $false
if ($args.Count -gt 0 -and $args[0] -eq "--emit-gh-output") {
    $emitGhOutput = $true
}

$foundTag = $true
$lastTag = git tag --merged HEAD --sort=-version:refname --list "v[0-9]*.[0-9]*.[0-9]*" | Select-Object -First 1
if (-not $lastTag) {
    $foundTag = $false
    $lastTag = "1.0.0"
}

function Normalize-Version {
    param([string]$Version)
    return $Version -replace '^v', ''
}

$currVer = Normalize-Version -Version $lastTag

$subjects = @()
if ($foundTag) {
    $range = "$lastTag..HEAD"
    $subjects = git log --format=%s $range 2>$null
    if (-not $subjects) {
        $subjects = @()
    }
} else {
    $subjects = git log --first-parent --format=%s HEAD 2>$null
    if (-not $subjects) {
        $subjects = @()
    }
}

if ($subjects -is [string]) {
    $subjects = @($subjects)
}

if ($subjects.Count -eq 0) {
    $bump = "none"
} else {
    $bump = "patch"

    foreach ($subject in $subjects) {
        if ($subject -match '^major(\([^)]+\))?:') {
            $bump = "major"
            break
        }
    }

    if ($bump -eq "patch") {
        foreach ($subject in $subjects) {
            if ($subject -match '^feat(\([^)]+\))?:') {
                $bump = "minor"
                break
            }
        }
    }

    if ($bump -eq "patch") {
        foreach ($subject in $subjects) {
            if ($subject -match '^fix(\([^)]+\))?:') {
                $bump = "patch"
                break
            }
        }
    }
}

$versionParts = $currVer -split '\.'
$major = [int]$versionParts[0]
$minor = [int]$versionParts[1]
$patch = [int]$versionParts[2]

$nextVer = $currVer
switch ($bump) {
    "major" {
        $major++
        $minor = 0
        $patch = 0
        $nextVer = "$major.$minor.$patch"
    }
    "minor" {
        $minor++
        $patch = 0
        $nextVer = "$major.$minor.$patch"
    }
    "patch" {
        $patch++
        $nextVer = "$major.$minor.$patch"
    }
    "none" {
    }
}

$hasRelease = 0
if (-not $foundTag) {
    $hasRelease = 1
    $bump = "none"
    $nextVer = $currVer
} elseif ($bump -ne "none") {
    $hasRelease = 1
}

if ($emitGhOutput) {
    $outputFile = $env:GITHUB_OUTPUT
    Add-Content -Path $outputFile -Value "last_tag=$currVer"
    Add-Content -Path $outputFile -Value "next_tag=v$nextVer"
    Add-Content -Path $outputFile -Value "bump=$bump"
    Add-Content -Path $outputFile -Value "has_release=$hasRelease"
} else {
    Write-Output "LAST_TAG=$currVer"
    Write-Output "NEXT_TAG=v$nextVer"
    Write-Output "BUMP=$bump"
    Write-Output "HAS_RELEASE=$hasRelease"
}
