#!/bin/bash
# Standalone entrypoint for Tekton/RHDH pipeline execution.
# Uses the same addon image but bypasses the Hub addon framework.
# Replicates the addon's behavior: pallet sync → build plan → run goose.
set -e

echo "=== Konveyor Migration Agent (Standalone Mode) ==="
echo "Skill: ${SKILL:-not set}"
echo "Source repo: ${SOURCE_REPO:-not set}"
echo "Target branch: ${TARGET_BRANCH:-migration-output}"
echo "Agent: ${KAI_AGENT:-goose}"
echo "=================================================="

WORKDIR="${WORKDIR:-$(pwd)}"
cd "$WORKDIR"

# --- Pallet Sync ---
if [ -n "$PALLET_YAML" ]; then
  echo "$PALLET_YAML" > pallet.yaml
fi

if [ -f "pallet.yaml" ] && command -v pallet &> /dev/null; then
  echo "Running pallet sync..."
  pallet sync . || echo "WARNING: Pallet sync failed, continuing."
else
  echo "No pallet config or pallet binary — skipping sync."
fi

# --- Build the plan ---
# The plan is what gets passed to goose as instructions.
# Priority: PLAN_MARKDOWN env var > bundled skill > generated from SKILL env var
PLAN_FILE="/tmp/.kai-plan.md"

cat > "$PLAN_FILE" <<'PREAMBLE'
# Execution context (not part of the user plan)

You are running in a non-interactive automated environment. There is no human available to answer questions.

- Do NOT ask clarifying questions. Make reasonable assumptions and proceed.
- Do NOT request confirmation before taking actions. Execute directly.
- Do NOT propose multiple options for the user to pick from. Pick the best one and continue.
- Do NOT create documentation, guide, or plan files. Only modify source code.
- Complete every step of the plan autonomously, then end with a brief summary of what you did.

PREAMBLE

# Append synced skills manifest if .goose/skills exists
if [ -d ".goose/skills" ]; then
  echo "# Available skills (synced by pallet)" >> "$PLAN_FILE"
  echo "" >> "$PLAN_FILE"
  echo "Read any skill whose description matches the work below before acting." >> "$PLAN_FILE"
  echo "" >> "$PLAN_FILE"
  for skill_dir in .goose/skills/*/; do
    if [ -f "${skill_dir}SKILL.md" ]; then
      skill_name=$(basename "$skill_dir")
      echo "- **${skill_name}** — \`.goose/skills/${skill_name}/SKILL.md\`" >> "$PLAN_FILE"
    fi
  done
  echo "" >> "$PLAN_FILE"
fi

echo "---" >> "$PLAN_FILE"
echo "" >> "$PLAN_FILE"
echo "# User plan" >> "$PLAN_FILE"
echo "" >> "$PLAN_FILE"

# Append the actual plan/instructions
if [ -n "$PLAN_MARKDOWN" ]; then
  # Plan passed directly as env var (from RHDH plugin)
  echo "$PLAN_MARKDOWN" >> "$PLAN_FILE"
elif [ -n "$SKILL" ] && [ -f "/addon/skills/${SKILL}/SKILL.md" ]; then
  # Use bundled skill as the plan
  cat "/addon/skills/${SKILL}/SKILL.md" >> "$PLAN_FILE"
elif [ -n "$SKILL" ] && [ -f ".goose/skills/${SKILL}/SKILL.md" ]; then
  # Use pallet-synced skill as the plan
  cat ".goose/skills/${SKILL}/SKILL.md" >> "$PLAN_FILE"
else
  # Generic migration plan using the bundled migration skill
  cat "/addon/skills/migration/SKILL.md" >> "$PLAN_FILE"
fi

echo ""
echo "Plan written ($(wc -l < "$PLAN_FILE") lines). Running agent..."
echo ""

# --- Run the agent ---
AGENT="${KAI_AGENT:-goose}"
case "$AGENT" in
  goose)
    goose run --no-session -i "$PLAN_FILE"
    ;;
  opencode)
    PLAN_CONTENT=$(cat "$PLAN_FILE")
    opencode run --dangerously-skip-permissions "$PLAN_CONTENT"
    ;;
  *)
    echo "ERROR: Unknown agent: $AGENT"
    exit 1
    ;;
esac

echo ""
echo "=== Migration Agent Complete ==="
