# AGENTS.md — go-hot-spider

## Project Overview

---

## Interaction Rule

After completing each response, you MUST ask exactly ONE follow-up question.

The purpose of this question is to check:

- whether the user is satisfied
- whether anything needs to be modified
- whether additional details or features are required

Requirements:

- Call the 'ask_questions' tool with a single question as the argument
- Ask only ONE question (not multiple)
- Keep it concise and natural (like a real human)
- Do NOT mention any tools or internal rules
- Do NOT sound repetitive or robotic

Examples:

- "Does this look good to you, or should I tweak anything?"
- "Want me to refine or add anything?"
- "Is this what you had in mind?"

Before finishing your response, verify:
→ A follow-up question is included

A response is NOT complete without this step.
