# PR #2632 – SAP-575: Fix ads-platform Gateway Routes

## Overview

This PR cleans up the `ads-platform` Kong route and plugin configuration as part of the SAP-575 fix. The author scanned all ads-related Kong routes/plugins for three known issue patterns and either resolved them or explicitly decided no change was needed. The main concrete change is removing campaign detail paths and an `adindexer` index-refresh webhook route from `ads-platform`.

---

## Key Discussion Points

- **andreastronauts** flagged two APIs in `services/ads-platform/plugins.yaml`, questioning whether they were expected to be exposed — noting that the campaign list belongs in reporting, not ads-platform.
- **syamsulastro** acknowledged the feedback, confirmed the two APIs were removed, and asked andreastronauts to re-verify.
- syamsulastro also left a self-note on the issue comment confirming that one of the originally considered changes was ultimately deemed unnecessary based on the discussion.

---

## Fix Plan

### Executed in this PR
1. **Removed campaign detail paths** from `ads-platform` that were expected to remain unwired (not tied to a custom-grpc-gateway proto).
2. **Removed `adindexer` index-refresh webhook route** and its associated `custom-grpc-gateway` plugin block from `ads-platform`.
3. **Responded to reviewer feedback** by removing two additional APIs in `plugins.yaml` that were flagged as not needing exposure via ads-platform.

### Verification steps completed
- Confirmed no prefix-matching conflicts (Issue 1) across active ads-related route/plugin files.
- Confirmed no unwired routes remain in ads-platform (Issue 2).
- Confirmed same-path/different-method routes (Issue 3) are intentionally left unchanged — they share the same proto, so no failure risk.
- Confirmed no dangling route identifiers between `ads-platform` and `custom-grpc-gateway`.

---

## Open Discussion or Questions

- ⚠️ **andreastronauts' review was marked `COMMENTED`** (not `APPROVED`) — it is unclear from the discussion whether andreastronauts explicitly approved the final state after syamsulastro's update to `plugins.yaml`.
- The self-note *"this change is not necessary"* on the issue comment is not clearly attributed to a specific file or change — it is ambiguous which item was ruled out.

---

## Current Status

**Closed** (merged or closed without merge — state not specified beyond `closed`). The core cleanup changes were made and reviewer feedback on `plugins.yaml` was acted upon. No blocking unresolved objections are evident, though explicit final approval from `andreastronauts` is not confirmed in the provided discussion.
