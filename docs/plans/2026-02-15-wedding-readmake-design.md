# Wedding App Documentation & Makefile Design

**Date:** 2026-02-15
**Status:** Approved

## Overview

Create comprehensive documentation and Makefile integration for the wedding wishes application located at `frontend-app/wedding/`.

## Project Context

The wedding app is a full-stack wedding invitation application:
- **Frontend:** Static HTML wedding invitation with photo gallery and background music
- **Backend:** Express.js server with REST API for guestbook wishes
- **Data Storage:** File-based storage in `whises.txt`
- **Gallery:** 5 photos with dedicated viewer page

## Deliverables

### 1. README.md

Location: `frontend-app/wedding/README.md`

**Sections:**
1. Overview - App description
2. Features - Photo gallery, music, guestbook
3. Quick Start - One-command startup via Makefile
4. Customization - How users can modify photos, text, music
5. Project Structure - File/folder organization
6. API Endpoints - GET/POST /api/wishes
7. Deployment - Hosting suggestions

### 2. Root Makefile Additions

Location: `/Users/m.syamsularifin/go/portofolio/Makefile`

**New Targets:**
```makefile
WEDDING_DIR ?= frontend-app/wedding

.PHONY: wedding-start wedding-stop

wedding-start:
	cd $(WEDDING_DIR) && npm start

wedding-stop:
	@echo "Stop the wedding server with Ctrl+C or kill the process"
```

**Integration Pattern:** Follows existing pattern (like `infra-start`, `swagger`, `up`, `stop`)

## Skills Analysis

| Skill | Purpose | Status |
|-------|---------|--------|
| `brainstorming` | Explore requirements | Complete |
| `writing-plans` | Create implementation plan | Next |
| `verification-before-completion` | Verify README & Makefile work | Pending |

## Design Decisions

1. **Approach 1 (Simple):** Add targets directly to root Makefile
   - Rationale: Consistent with existing pattern, minimal complexity

2. **Comprehensive README:** Target both developers and end-users
   - Rationale: App may be used/customized by non-technical users

3. **Wishes Data Handling:** Commit `whises.txt` once, then ignore changes
   - Implementation: Add `frontend-app/wedding/whises.txt` to `.gitignore`

## Implementation Plan

To be created by `writing-plans` skill.
