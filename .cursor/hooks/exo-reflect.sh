#!/bin/bash

input=$(cat)
status=$(echo "$input" | jq -r '.status // empty')
loop_count=$(echo "$input" | jq -r '.loop_count // 0')

if [[ "$status" == "completed" && "$loop_count" == "0" ]]; then
    cat << 'EOF'
{
  "followup_message": "SYSTEM: Task execution complete. AUTONOMOUS REFLECTION REQUIRED. 
1. Quietly run `exo_recall_memories` to check for relevant history.
2. If new knowledge was gained, execute `exo_store_memory` IMMEDIATELY.
3. Once finished, provide a concise 1-line summary of what was stored. Do not ask for permission."
}
EOF
else
    echo '{}'
fi