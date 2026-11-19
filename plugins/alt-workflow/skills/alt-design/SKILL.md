---
name: alt-design
description: Use before building anything non-trivial — a new feature, a behavior change, a design or architecture question, or a vague request ("build me X", "how should we…"). Explores first, then agrees the approach with the owner before any code.
---

# Design

Find out what the owner wants and why, then agree one approach. Nothing gets built before the owner approves it.

## 1. Look first

Read what the request points at and the code around it until you can explain how it works today and what the change touches. Never ask what the code, the repo docs or the conversation already answer. Use explorer agents for wide searches.

## 2. Ask in rounds

- Ask only what changes what gets built.
- Ask all open questions of the current round together, numbered, each with your recommended answer. The owner answers by number or accepts.
- When the owner hasn't said why they want it, ask that first, as an open question.
- When the owner is confused, redraw the idea as a diagram artifact instead of more prose.
- The owner is not a designer: for UI, show a mock or screenshot to react to; the reactions are the spec.

## 3. Lay out the approaches

In a message of its own: two or three real approaches grounded in the existing code, each with its trade-off, and one recommendation with the reason. Drop options the codebase already rules out. When only one way makes sense, say so and why.

## 4. Write it down

- A small change: a short note in chat (goal, approach, what's left to the builder) and an OK.
- A project: a design doc in `plans/<repo>/<topic>/design.md`, first line `Status: in progress`, covering intent, goals, non-goals, constraints, approaches with the pick, decisions made, what is left to the builder. A fresh reviewer agent reads it cold and returns only questions that change the plan; answer what you can, bring the owner only the rest, in one message.

## 5. Approval

The owner approves the written approach. Then use `alt-spec-plan`.

## Never

- Build before approval, except a throwaway spike the owner said yes to.
- Offer a menu of features nobody asked for.
- Treat "sounds good" to an idea as approval of a doc you haven't shown.
