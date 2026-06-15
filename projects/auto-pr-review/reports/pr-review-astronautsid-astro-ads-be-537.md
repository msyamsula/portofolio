# PR #537: sap-490 — Add Display Explain Endpoints

## Overview

This PR adds two internal gRPC diagnostic endpoints (`ExplainSearchAds`, `ExplainBrowseAds`) to the display domain. They allow engineers to inspect *why* a specific ad appeared or didn't appear in search/browse results, providing a structured breakdown of score components, bid source, relevance matching, and negative-case diagnostics. The feature includes an ES repository layer, service pipeline, gRPC handlers, observability metrics, rate limiting (1 RPS shared), and a design document.

---

## Key Discussion Points

### AI Reviewer (Claude) — Initial Pass
Three **HIGH** issues were flagged as blocking merge:

1. **gRPC status codes leaking into service layer** (`explain_limiter.go`): The service imports `grpc/codes` and returns a `ResourceExhausted` gRPC error directly, forcing the handler to reverse-engineer the code via `isRateLimitError`. Convention requires service layer to return a domain error (e.g., `ErrExplainRateLimited`), with the handler doing the mapping.

2. **Wrong ES document ID format in negative-case diagnosis** (`explain_search_ads.go`, `explain_browse_ads.go`): Code uses only `adID` for the `docID`, but the design doc specifies `"<ad_id>_<product_id>"`. This means every `ExplainDocument` call returns `found: false`, making all negative cases incorrectly report `not_indexed`.

Two **MEDIUM** issues (non-blocking):
- Category boost missing from `buildBoostingScore` — `BoostMaps.Category` is populated but never consumed, making `BoostingScore.Total` and `FinalScore` lower than actual ES values.
- Metric tag constants embed key+value (`"type:browse"`) instead of just the key (`"type:"`), violating observability conventions.

Three **LOW** issues (non-blocking): unnecessary `_ = ctx`, silently discarded parameters in `toExplainHit`, unrelated `admanager` entity refactoring included in the PR.

### Copilot Reviewer — First Pass
Five comments, all **Medium** severity:

1. **`sort` key dropped in `buildExplainSearchBody`**: Only `query` is forwarded from the original ES body; `sort`, `collapse`, and `aggs` are dropped. Without `sort`, ES uses default `_score` sorting and `hit.Sort` is always empty, so `SortScore` is always 0 — misleading since docs describe it as the "authoritative final_score from ES."
2. **Missing `Size` validation** (both search and browse explain): `SearchAds`/`BrowseAds` guard against `Size > MaxPageSize` (RULE-DSP-20); explain endpoints skip this check.
3. **Missing test files**: Design doc lists `explain_search_ads_test.go`, `explain_browse_ads_test.go`, `explain_score_test.go`, `explain_parser_test.go` as deliverables; none were present at initial review.
4. **Log level inconsistency**: Sort parse failure uses `logger.Log().Info` instead of `Warn`, deviating from the pattern in the same file.

### Author Responses (after June 8 push — commit `7dd3e261`)

| Issue | Author Response |
|---|---|
| `sort` key dropped | **Fixed** — `buildExplainSearchBody` now forwards `original["sort"]`; covered by `TestBuildExplainSearchBody_ForwardsSortAndFiltersAd` |
| Log level inconsistency | **Fixed** — removed the log entirely; sort parse is now best-effort (`_ = json.Unmarshal`) defaulting `SortScore` to 0 |
| Missing tests | **Fixed** — added `explain_score_test.go`, `explain_parser_test.go`, `explain_search_ads_test.go`, `explain_browse_ads_test.go` |
| `Size` validation (both endpoints) | **Resolved as non-issue** — `buildExplainSearchBody` hardcodes `"size": 1` and discards the request `Size`, so RULE-DSP-20 risk doesn't apply here |

### Copilot Reviewer — Second Pass (post-author fixes)

Four new issues raised:

1. **`buildBidBreakdown` misclassifies `TARGETING_MANUAL` ads**: Returns `Strategy: "AUTO"` whenever `TopKeyword` is nil — even for manual ads that simply didn't match a keyword. Design doc Case 2 specifies `targeting_type: "TARGETING_MANUAL"` and `bid_source: "ad_level_fallback"` for this scenario. The downstream mapper `toBidBreakdownProto` then also derives the wrong values.
2. **`firstNonZero` function name is misleading**: Returns the first element regardless of value (including zero). Suggested rename to `firstOrZero`.
3. **Doc/code mismatch on error code**: Design doc says `ErrExplainNotConfigured` → `codes.FailedPrecondition`; handler uses `codes.Unimplemented`.
4. **`buildSearchNegativeResult` / `buildBrowseNegativeResult` duplication**: ~40 lines of near-identical code differing only in metric tag and log message; could be consolidated into one method with a `placement string` param.

### Author Responses (June 8, second round)

| Issue | Author Response |
|---|---|
| Negative result method duplication | **Intentional** — duplicate is kept by design |
| PR description / `codes.Unimplemented` mismatch | Acknowledged as a description mismatch after review |

---

## Fix Plan

### Confirmed Fixes Needed (unresolved as of last review)

#### 1. `buildBidBreakdown` misclassifies manual ads without keyword match
**Steps:**
1. In `explain_bid.go`, add a parameter (or access) for the ad's actual targeting type.
2. Set `Strategy` based on the ad's targeting type, not on whether `TopKeyword` is nil.
3. For `TARGETING_MANUAL` + no keyword match: set `Strategy: "MANUAL"`, `BidSource: "ad_level_fallback"`.
4. Update `toBidBreakdownProto` in `explain_mapper.go` to correctly map `"ad_level_fallback"` for this case.
5. Add test case `TestBuildBidBreakdown_ManualWithoutKeyword_StillManual` to `explain_score_test.go`.

#### 2. Design doc / handler error code mismatch
**Steps:**
1. In `docs/etc/display-explain.md`, update the error mapping table to reflect `ErrExplainNotConfigured` → `codes.Unimplemented` (or change the handler to `codes.FailedPrecondition` — pick one and align both).

#### 3. HIGH issues from Claude (status unknown — no author acknowledgment on record)
- **gRPC codes in service layer** (`explain_limiter.go`): Define `ErrExplainRateLimited` in `entity/explain.go`; service returns domain error; handler maps to `codes.ResourceExhausted`.
- **Wrong docID format** (both `explain_search_ads.go` and `explain_browse_ads.go`): Use `fmt.Sprintf("%d_%d", adID, productID)` and thread `productID` through `buildSearchNegativeResult` / `buildBrowseNegativeResult`.

> ⚠️ The author's commit `7dd3e261` is referenced for the Copilot fixes, but no explicit acknowledgment of the Claude HIGH issues appears in the discussion.

---

## Open Discussion / Unresolved Questions

1. **Were the three Claude HIGH issues actually fixed in `7dd3e261`?** The author addressed Copilot's comments explicitly but made no public comment on: gRPC codes leaking into service layer, incorrect docID format (`adID` only vs. `adID_productID`), and missing category boost in score reconstruction. It is unclear whether these were addressed in the commit silently or remain open.

2. **`firstNonZero` rename**: Copilot flagged the misleading function name; no author response recorded. Minor but could cause confusion for future maintainers.

3. **Duplicate negative-result methods**: Author explicitly said duplication is intentional. The reviewer's consolidation suggestion remains unaccepted
