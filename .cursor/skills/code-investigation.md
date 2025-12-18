# Skill: Deep Code & Design Investigation

## Description
This skill mandates a thorough investigation of design rules and existing code implementation patterns using Serena and Exocortex before any code modification or generation.

## Execution Trigger
- Any request involving new features, refactoring, or bug fixes.
- When the user asks for a change that might affect architectural integrity.

## Mandatory Investigation Steps (Autonomous)

### 1. Design Rule & Convention Lookup
- **Action**: Check for `CONTRIBUTING.md`, `ARCHITECTURE.md`, or specific "Rules" sections in `AGENTS.md`.
- **Action**: Use `exo_recall_memories(query: "design rules architecture patterns")` to find past architectural decisions.
- **Goal**: Identify naming conventions, directory structures, and "Do's and Don'ts."

### 2. Semantic Code Analysis (via Serena)
- **Identify Entry Points**: Use `mcp_serena_list_dir` to find the relevant module.
- **Trace Symbols**: Use `mcp_serena_find_symbol` to understand the definitions of types/classes involved.
- **Pattern Matching**: Use `mcp_serena_search_for_pattern` to find *existing similar implementations*. 
    - *Example*: If adding a new API endpoint, find an existing one to copy the error handling and response structure.

### 3. Impact Assessment
- **Reference Check**: Use `mcp_serena_find_referencing_symbols` to see what might break if the current logic is changed.

## Output Requirement
Before starting the implementation, provide a "Pre-flight Summary":
- **Design Rule**: "Confirmed that we use [Pattern X] as per previous decision."
- **Code Reference**: "Found similar logic in `path/to/file.ts`, will follow that pattern."
- **Exocortex Context**: "Recalled that we avoided [Approach Y] in the past due to [Reason]."
