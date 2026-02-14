# Wedding App Documentation & Makefile Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create comprehensive README.md documentation and add Makefile targets for the wedding wishes application.

**Architecture:**
- README.md: Comprehensive documentation serving both developers and end-users
- Makefile: Simple addition of wedding-start/wedding-stop targets to root Makefile following existing patterns

**Tech Stack:** Markdown documentation, GNU Make, Express.js, Node.js

---

## Task 1: Create README.md for Wedding App

**Files:**
- Create: `frontend-app/wedding/README.md`

**Step 1: Create README.md with complete documentation**

```markdown
# Wedding Wishes Application

A beautiful wedding invitation website with a guestbook feature for collecting wishes from guests.

## Features

- **Elegant Wedding Invitation** - Responsive single-page invitation with forest green theme
- **Photo Gallery** - Interactive gallery viewer with 5 couple photos
- **Background Music** - Ambient audio for enhanced experience
- **Guestbook Wishes** - API-driven guestbook for collecting and displaying guest messages
- **Mobile Responsive** - Optimized for all screen sizes

## Quick Start

Start the wedding application server:

```bash
# From root directory
make wedding-start
```

Or manually:

```bash
cd frontend-app/wedding
npm install
npm start
```

The server will start at http://localhost:3000

## Project Structure

```
wedding/
├── index.html           # Main wedding invitation page
├── gallery-viewer.html  # Photo gallery viewer
├── server.js            # Express server for API
├── package.json         # Node.js dependencies
├── whises.txt           # Guest wishes storage (file-based)
├── audio.mp3            # Background music
├── gallery/             # Photo gallery images (5 JPEGs)
│   ├── 1.jpeg
│   ├── 2.jpeg
│   ├── 3.jpeg
│   ├── 4.jpeg
│   └── 5.jpeg
└── images/              # UI assets (bank logos, wax seal)
```

## For Users: Customization Guide

### Changing Photos

Replace photos in the `gallery/` folder:
- Keep filenames as `1.jpeg`, `2.jpeg`, `3.jpeg`, `4.jpeg`, `5.jpeg`
- Recommended size: 800x1200px (portrait orientation)
- Format: JPEG

### Changing Text

Edit text directly in `index.html`:
- Couple names
- Event date and location
- Welcome messages

### Changing Music

Replace `audio.mp3` with your preferred background music.

### Changing Colors

Edit the color variables in `index.html`:
```javascript
colors: {
    primary: "#1b3022",    // Forest green
    gold: "#d4a017",       // Gold accent
}
```

## For Developers

### Installation

```bash
cd frontend-app/wedding
npm install
```

### API Endpoints

**GET /api/wishes**
- Returns all guest wishes
- Response: `{ "wishes": [{ "name": "John", "message": "Congratulations!" }] }`

**POST /api/wishes**
- Submit a new guest wish
- Body: `{ "name": "John", "message": "Best wishes!" }`
- Returns updated wishes list

### Data Storage

Guest wishes are stored in `whises.txt` (one wish per line):
```
Name: Message here
Another Name: Another message
```

### Development

The server uses Express.js with:
- Static file serving for HTML/CSS/JS
- JSON body parsing
- File-based data storage
- Port 3000 (configurable via PORT env var)

## Deployment

### Static Hosting (Frontend Only)

For static hosting (GitHub Pages, Netlify, Vercel):
1. Remove the server.js dependency
2. Pre-populate wishes or use external API
3. Deploy `index.html` and assets

### Full Stack Hosting

For platforms like Render, Railway, Fly.io:
1. Push code to GitHub
2. Connect repository to platform
3. Set start command to `npm start` (runs `node server.js`)
4. Ensure PORT environment variable is respected

### Environment Variables

- `PORT` - Server port (default: 3000)

## License

Private project.
```

**Step 2: Verify README.md was created**

Run: `ls -la frontend-app/wedding/README.md`
Expected: File exists with content

**Step 3: Commit README.md**

```bash
git add frontend-app/wedding/README.md
git commit -m "docs: add comprehensive README for wedding app"
```

---

## Task 2: Update Root .gitignore for whises.txt

**Files:**
- Modify: `/Users/m.syamsularifin/go/portofolio/.gitignore`

**Step 1: Add whises.txt to .gitignore**

Add this line to the existing `.gitignore`:

```gitignore
# Wedding app guest wishes (user data)
frontend-app/wedding/whises.txt
```

**Step 2: Verify .gitignore was updated**

Run: `cat .gitignore | grep whises`
Expected: Line with `frontend-app/wedding/whises.txt`

**Step 3: Commit .gitignore changes**

```bash
git add .gitignore
git commit -m "chore: ignore wedding app whises.txt changes"
```

---

## Task 3: Add Wedding Targets to Root Makefile

**Files:**
- Modify: `/Users/m.syamsularifin/go/portofolio/Makefile`

**Step 1: Read current Makefile to understand exact structure**

Run: `cat /Users/m.syamsularifin/go/portofolio/Makefile`
Expected: See existing targets like `infra-start`, `infra-stop`, `swagger`, `up`, `stop`

**Step 2: Add wedding directory variable and targets**

Add after the existing variables (line 3-4):

```makefile
WEDDING_DIR ?= frontend-app/wedding
```

Add to the `.PHONY` declaration (after `stop`):

```makefile
.PHONY: infra-start infra-stop swagger up stop wedding-start wedding-stop
```

Add new targets at the end of the file:

```makefile

wedding-start:
	cd $(WEDDING_DIR) && npm start

wedding-stop:
	@echo "Stop the wedding server with Ctrl+C or kill the process"
```

**Step 3: Verify Makefile syntax**

Run: `make -n wedding-start`
Expected: Shows command `cd frontend-app/wedding && npm start` without executing

**Step 4: Test wedding-start target (dry-run verification)**

Run: `cd frontend-app/wedding && test -f package.json && echo "package.json exists"`
Expected: "package.json exists"

**Step 5: Commit Makefile changes**

```bash
git add Makefile
git commit -m "feat: add wedding-start and wedding-stop Makefile targets"
```

---

## Task 4: Update Root README (if exists) to Document Wedding Targets

**Files:**
- Modify: `/Users/m.syamsularifin/go/portofolio/README.md` (if exists, otherwise skip)

**Step 1: Check if root README exists**

Run: `test -f /Users/m.syamsularifin/go/portofolio/README.md && echo "exists" || echo "not found"`
Expected: If "not found", skip this task

**Step 2: If exists, add wedding section to Makefile targets documentation**

Add to existing Makefile section:

```markdown
### Wedding App

- `make wedding-start` - Start the wedding wishes application server
- `make wedding-stop` - Display instructions to stop the server
```

**Step 3: Commit README changes (if modified)**

```bash
git add README.md
git commit -m "docs: document wedding Makefile targets in root README"
```

---

## Task 5: Final Verification

**Files:**
- Test: `frontend-app/wedding/`
- Test: `/Users/m.syamsularifin/go/portofolio/Makefile`

**Step 1: Verify wedding-start target syntax**

Run: `cd /Users/m.syamsularifin/go/portofolio && make -n wedding-start`
Expected: Shows `cd frontend-app/wedding && npm start`

**Step 2: Verify README.md exists and is readable**

Run: `head -20 frontend-app/wedding/README.md`
Expected: Shows first 20 lines of README with title

**Step 3: Verify .gitignore includes whises.txt**

Run: `grep whises .gitignore`
Expected: Shows `frontend-app/wedding/whises.txt`

**Step 4: Create summary of changes**

Run: `git log --oneline -4`
Expected: Shows recent commits for this feature

**Step 5: Final commit (if any remaining changes)**

```bash
git status
# (if any uncommitted changes remain)
git add .
git commit -m "chore: final cleanup for wedding documentation"
```

---

## Completion Criteria

- [ ] README.md created in wedding folder with all sections
- [ ] .gitignore updated to ignore whises.txt changes
- [ ] Makefile has wedding-start and wedding-stop targets
- [ ] All changes committed to git
- [ ] `make wedding-start` shows correct command in dry-run mode
