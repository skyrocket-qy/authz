# HTTP Engine

## Implement
1. Use Url pattern to match and get resources
2. For resources, check on zanzibar

## Feature
to check can visit /A/B, must check: 1. it can visit /A/B or /A/* 2. it will not block /A/B or /A/* 3. not block /A

Static segments (/A/B/C)

Variable parameters (/A/:id/C)

Wildcards (/A/*)

Method matching (GET, POST, etc.)

🔹 Approve Rule (allow / route to handler)

Normal use case: define a route and its handler.

Example:

/A/B → handler1
/A/B/C → handler2


✅ If request matches, it’s “approved” and processed.

🔹 Reject Rule (block explicitly)

Sometimes you want to reject certain paths, even if a handler exists or a wildcard could match.

Example:

/A/* → handlerWildcard
/A/B → reject


/A/X → goes to handlerWildcard.

/A/B → ❌ explicitly blocked (maybe return 403 Forbidden).
🔹 Default Rule

If no explicit approve/reject rule matches → default reject (404) in most routing engines.

But some engines allow a fallback approve (e.g., proxy to another service).

🔹 Priority Order

When both approve and reject rules exist:

Exact reject > exact approve > wildcard approve > default reject

Ensures a block rule always wins over an allow.