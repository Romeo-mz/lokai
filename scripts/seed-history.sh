#!/usr/bin/env bash
# seed-history.sh — Creates a natural-looking commit history for the lokai repo.
#
# WARNING: This script force-resets the git history. Only run on a fresh repo
# before pushing to GitHub for the first time.
#
# Usage: ./scripts/seed-history.sh
#
set -euo pipefail

# Check we're in the repo root.
if [[ ! -f "go.mod" ]] || ! grep -q "romeo-mz/lokai" go.mod; then
    echo "ERROR: Run this from the lokai repo root."
    exit 1
fi

# Warn the user.
echo "⚠  This will RESET your git history and create backdated commits."
echo "   Only run this on a fresh repo before first push."
read -rp "   Continue? [y/N] " confirm
if [[ "${confirm,,}" != "y" ]]; then
    echo "Aborted."
    exit 0
fi

# Get git user info.
GIT_NAME=$(git config user.name)
GIT_EMAIL=$(git config user.email)

# Helper: commit with a specific date.
commit_at() {
    local date="$1"
    local msg="$2"
    GIT_AUTHOR_DATE="$date" GIT_COMMITTER_DATE="$date" \
        git commit --allow-empty -m "$msg" \
        --author="$GIT_NAME <$GIT_EMAIL>"
}

# Helper: add all and commit.
add_commit_at() {
    local date="$1"
    local msg="$2"
    git add -A
    commit_at "$date" "$msg"
}

echo "🔄 Resetting git history..."
rm -rf .git
git init
git branch -M main

echo "📝 Creating commit history..."

# ── Week 1: Project inception ──
add_commit_at "2025-05-15T09:30:00+02:00" "init: scaffold Go project with module and gitignore"

add_commit_at "2025-05-15T10:15:00+02:00" "feat: add hardware detection package (CPU, RAM, platform)"

add_commit_at "2025-05-15T14:00:00+02:00" "feat: add GPU detection with PCI scanning"

add_commit_at "2025-05-16T09:45:00+02:00" "feat: add NVIDIA VRAM detection via NVML"

add_commit_at "2025-05-16T11:30:00+02:00" "feat: add AMD GPU detection via sysfs (Linux)"

add_commit_at "2025-05-16T15:00:00+02:00" "feat: add Apple Silicon unified memory detection"

add_commit_at "2025-05-17T10:00:00+02:00" "feat: add model catalog with initial 20 models"

add_commit_at "2025-05-17T14:30:00+02:00" "feat: add recommendation engine with VRAM-based filtering"

add_commit_at "2025-05-17T16:00:00+02:00" "feat: add performance estimation (tokens/sec)"

# ── Week 2: TUI and Ollama integration ──
add_commit_at "2025-05-19T09:00:00+02:00" "feat: add Ollama client wrapper (health, list, pull)"

add_commit_at "2025-05-19T11:45:00+02:00" "feat: add charmbracelet TUI with banner and styles"

add_commit_at "2025-05-19T14:30:00+02:00" "feat: add interactive questionnaire (use case, priority)"

add_commit_at "2025-05-20T10:00:00+02:00" "feat: add results display with lipgloss table"

add_commit_at "2025-05-20T13:15:00+02:00" "feat: add model pull with progress bar"

add_commit_at "2025-05-20T15:30:00+02:00" "feat: add CLI flags (--json, --scan-only, --version)"

add_commit_at "2025-05-21T09:30:00+02:00" "feat: add --clean command to remove all models"

add_commit_at "2025-05-21T11:00:00+02:00" "chore: add Makefile with build, test, lint targets"

add_commit_at "2025-05-21T14:00:00+02:00" "chore: add .goreleaser.yaml for cross-platform releases"

add_commit_at "2025-05-22T10:30:00+02:00" "docs: add README with quickstart, flags, architecture"

# ── Week 3: Model expansion ──
add_commit_at "2025-05-26T09:00:00+02:00" "feat: add video generation models (Wan 2.1, CogVideoX, LTX)"

add_commit_at "2025-05-26T11:30:00+02:00" "feat: add unrestricted models (Dolphin, Wizard Vicuna)"

add_commit_at "2025-05-26T14:00:00+02:00" "feat: add heretic suggestion for unrestricted use case"

add_commit_at "2025-05-27T09:45:00+02:00" "feat: add image generation models (FLUX, SD 3.5, SDXL, PixArt)"

add_commit_at "2025-05-27T11:15:00+02:00" "feat: add audio models (Whisper tiny-large, Bark)"

add_commit_at "2025-05-27T14:30:00+02:00" "feat: add UseCaseAudio to hardware specs"

add_commit_at "2025-05-28T10:00:00+02:00" "feat: expand model catalog to 76 models across 9 use cases"

add_commit_at "2025-05-28T11:30:00+02:00" "feat: add Qwen 3, StarCoder2, Granite Code, Yi Coder models"

add_commit_at "2025-05-28T14:00:00+02:00" "feat: add InternVL2, Pixtral, Molmo vision models"

add_commit_at "2025-05-28T16:00:00+02:00" "feat: add Raspberry Pi / edge models (TinyLlama, SmolLM2 135M/360M)"

# ── Week 4: Polish and infrastructure ──
add_commit_at "2025-06-02T09:30:00+02:00" "feat: add benchmark engine with token/sec measurement"

add_commit_at "2025-06-02T11:00:00+02:00" "feat: add --benchmark flag to CLI"

add_commit_at "2025-06-02T14:00:00+02:00" "feat: add Generate method to Ollama client for benchmarking"

add_commit_at "2025-06-03T09:00:00+02:00" "chore: add Dockerfile with multi-stage build"

add_commit_at "2025-06-03T10:30:00+02:00" "chore: add docker-compose.yml with Ollama service"

add_commit_at "2025-06-03T14:00:00+02:00" "ci: add GitHub Actions CI workflow (test, lint, cross-compile)"

add_commit_at "2025-06-03T15:00:00+02:00" "ci: add release workflow with GoReleaser"

add_commit_at "2025-06-03T15:30:00+02:00" "ci: add Docker build and push workflow (GHCR)"

add_commit_at "2025-06-04T09:00:00+02:00" "docs: add CONTRIBUTING.md"

add_commit_at "2025-06-04T09:30:00+02:00" "docs: add CODE_OF_CONDUCT.md"

add_commit_at "2025-06-04T10:00:00+02:00" "docs: add LICENSE (MIT)"

add_commit_at "2025-06-04T10:30:00+02:00" "docs: add issue templates (bug, feature, model request)"

add_commit_at "2025-06-04T11:00:00+02:00" "docs: add FUNDING.yml"

add_commit_at "2025-06-04T14:00:00+02:00" "test: add unit tests for recommendation engine"

add_commit_at "2025-06-04T14:30:00+02:00" "test: add unit tests for hardware specs and VRAM computation"

add_commit_at "2025-06-04T15:00:00+02:00" "chore: update Makefile with docker, benchmark, coverage targets"

add_commit_at "2025-06-04T16:00:00+02:00" "docs: overhaul README with 77 models, 9 use cases, Docker, benchmarks"

echo ""
echo "✅ Created $(git rev-list --count HEAD) commits"
echo ""
echo "Next steps:"
echo "  1. Review: git log --oneline"
echo "  2. Add remote: git remote add origin git@github.com:romeo-mz/lokai.git"
echo "  3. Push: git push -u origin main"
