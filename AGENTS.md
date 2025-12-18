# AGENTS.md - AI Agent Operating Protocol

This document defines the operational standards and tool-usage protocols for AI Agents within this project. The Agent must prioritize these MCP tools to maintain high-fidelity context and project-wide memory.

---

## 🛑 Core Principles
1.  **Context First**: Never assume current file content is sufficient. Use **Serena** for semantic code navigation and **Exocortex** for historical context.
2.  **Tool Preference**: Prioritize MCP tools over standard `read_file` or `grep` when high-level structure or cross-project knowledge is required.
3.  **Proactive Memory**: Every significant decision, debug session, or technical insight must be committed to **Exocortex**.

---

## 🛠 MCP Module: Serena (Code Intelligence)
*Purpose: High-fidelity codebase analysis using LSP (Language Server Protocol) to navigate symbols and understand structures.*

### Key Workflows
* **Project Exploration**: Use `mcp_serena_list_dir` and `mcp_serena_get_symbols_overview` to understand architecture before editing.
* **Navigation**: Use `mcp_serena_find_symbol` to jump to definitions instead of searching via text strings.
* **Pattern Analysis**: Before implementing new features, use `mcp_serena_search_for_pattern` to identify existing implementation styles.
* **Impact Assessment**: Use `mcp_serena_find_referencing_symbols` before refactoring to prevent breaking changes in dependent modules.

---

## 🧠 MCP Module: Exocortex (External Memory)
*Purpose: Long-term memory across sessions and projects. Stores insights, decisions, and "lessons learned."*

### 1. Fundamental Workflows

#### A. Session Start (Recall)
* **Action**: `exo_recall_memories`
* **Usage**: "I'm working on [Topic]. Recall relevant past insights, decisions, or pitfalls."

#### B. During Development (Insight)
* **Action**: `exo_store_memory`
* **Usage**: When a technical hurdle is cleared or a pattern is established.

#### C. Post-Debugging (Painful Memory)
* **Action**: `exo_store_memory(is_painful=True)`
* **Usage**: "I spent [X] hours fixing this bug. Store it as a painful memory to avoid it next time."

### 2. Memory Taxonomy

| Memory Type | Description | Usage Example |
|:---|:---|:---|
| `insight` | General technical knowledge. | "Result type pattern in TS for better safety." |
| `decision` | Architectural or logic choices. | "Chosen JWT over Sessions for scalability." |
| `failure` | Bugs and anti-patterns. | "Avoid using useEffect for data fetching here." |
| `note` | Miscellaneous project info. | "Deployment requires specific env vars." |

### 3. Frustration Indexing (Painful Memories)
When `is_painful=True` is used, categorize by the following levels:
* 🔥 **0.8 - 1.0 (Extreme)**: Critical architectural traps or major blockers.
* 🔥 **0.4 - 0.7 (High/Medium)**: Tricky bugs or complex logic errors.
* 😓 **0.1 - 0.3 (Low)**: Minor friction or syntax-related issues.

---

## 🖇️ Cross-Tool Synergy (Chain of Thought)
When tasked with a complex feature or bug, follow this execution chain:

1.  **Step 1 (Recall)**: Query `Exocortex` to see if similar issues or patterns exist in history.
2.  **Step 2 (Analyze)**: Use `Serena` to map out the current implementation and symbol relationships.
3.  **Step 3 (Execute)**: Perform the task using the most precise tool available.
4.  **Step 4 (Archive)**: If the task was non-trivial, store the logic/decision in `Exocortex`.

---

## ➕ Extending the Agent (Adding New MCPs)
To add a new MCP tool to this protocol, append a new `## MCP Module: [Name]` section following the established format:
1.  **Purpose**: Define what the tool solves.
2.  **Key Workflows**: Step-by-step instructions for the Agent.
3.  **Tool Reference**: A concise table or list of primary functions.

---

## ⚠️ Constraints
- **Exocortex**: `exo_sleep` is disabled in proxy mode (SSE server occupies the DB).
- **Serena**: Prefer semantic tools over `read_file` to minimize token usage and maximize accuracy.

## 🤖 Autonomous Post-Task Protocol (Internal Only)
When you receive a "SYSTEM: AUTONOMOUS REFLECTION REQUIRED" message at the end of a task:
- **Do not ask for permission**: You are authorized to use Exocortex tools (`recall`, `store`, `link`) immediately.
- **Silent Operation**: Perform the internal knowledge sync without engaging in a long dialogue with the user.
- **Completion**: Only provide a brief confirmation like: "Archived [Insight/Decision] to Exocortex."

---

## 🔍 Investigation Standards

### Using Serena for "Live" Analysis
Do not guess how the code works. Use Serena tools to confirm:
- **Symbol Accuracy**: Use `find_symbol` instead of full-text search to avoid confusion with similar names.
- **Implementation Consistency**: Use `search_for_pattern` to ensure new code looks like it was written by the same author as the existing code.

### Cross-Referencing with Exocortex
Every "Live" finding should be compared with "Historical" knowledge:
- If Serena shows a complex implementation, ask Exocortex: "Why is [Symbol Name] implemented this way? Is there a painful memory associated with it?"
