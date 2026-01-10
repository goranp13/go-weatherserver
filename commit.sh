#!/usr/bin/env bash
set -euo pipefail

# commit.sh - initialize git repo (if needed), commit changes, optionally add remote and push.
# Usage:
#  GIT_NAME="Your Name" GIT_EMAIL="you@example.com" REMOTE_URL="git@github.com:you/repo.git" ./commit.sh
# or set env vars in the environment and run:
#  ./commit.sh
# Optional: pass a custom commit message as the first argument.

COMMIT_MSG=${1:-"feat: extract frontend to templates/static, add README"}
GIT_NAME=${GIT_NAME:-}
GIT_EMAIL=${GIT_EMAIL:-}
REMOTE_URL=${REMOTE_URL:-}
BRANCH=${BRANCH:-main}

# Check if git is installed
if ! command -v git >/dev/null 2>&1; then
  echo "Error: git is not installed or not in PATH. Install git and re-run this script." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

# Initialize repo if not already a git repo
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Initializing new git repository..."
  git init
else
  echo "Git repository detected."
fi

# Configure user if provided
if [ -n "$GIT_NAME" ]; then
  git config user.name "$GIT_NAME"
  echo "Set git user.name to '$GIT_NAME'"
fi
if [ -n "$GIT_EMAIL" ]; then
  git config user.email "$GIT_EMAIL"
  echo "Set git user.email to '$GIT_EMAIL'"
fi

# Add everything and commit
git add -A
# If there are no changes to commit, exit gracefully
if git diff --cached --quiet; then
  echo "No changes to commit."
else
  git commit -m "$COMMIT_MSG"
  echo "Committed: $COMMIT_MSG"
fi

# Setup remote and push if REMOTE_URL is set
if [ -n "$REMOTE_URL" ]; then
  # If origin exists but points differently, warn and update
  if git remote get-url origin >/dev/null 2>&1; then
    existing=$(git remote get-url origin)
    if [ "$existing" != "$REMOTE_URL" ]; then
      echo "Remote 'origin' exists and points to: $existing"
      echo "Updating 'origin' to $REMOTE_URL"
      git remote remove origin
      git remote add origin "$REMOTE_URL"
    else
      echo "Remote 'origin' already set to $REMOTE_URL"
    fi
  else
    git remote add origin "$REMOTE_URL"
    echo "Added remote origin -> $REMOTE_URL"
  fi

  # Ensure branch exists locally
  if ! git show-ref --verify --quiet refs/heads/$BRANCH; then
    git branch -M $BRANCH || true
  fi

  echo "Pushing to origin/$BRANCH..."
  git push -u origin $BRANCH
  echo "Push completed."
else
  echo "REMOTE_URL not set. Skipping push. If you want to push, set REMOTE_URL and rerun."
fi

echo "Done."
