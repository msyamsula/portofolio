const host = "http://localhost:5000"
const themeToggle = document.getElementById('themeToggle');

function setTheme(theme) {
    const nextTheme = theme === 'dark' ? 'dark' : 'light';
    document.body.dataset.theme = nextTheme;
    themeToggle.textContent = nextTheme === 'dark' ? 'Light mode' : 'Dark mode';
    localStorage.setItem('theme', nextTheme);
}

const storedTheme = localStorage.getItem('theme');
const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
setTheme(storedTheme || (prefersDark ? 'dark' : 'light'));

themeToggle.addEventListener('click', () => {
    const currentTheme = document.body.dataset.theme;
    setTheme(currentTheme === 'dark' ? 'light' : 'dark');
});

function loadReadme() {
    const readmeEl = document.getElementById('readmeContent');
    const readmeZoomEl = document.getElementById('readmeZoom');
    if (!readmeEl || !readmeZoomEl) {
        return;
    }

    if (window.mermaid) {
        window.mermaid.initialize({ startOnLoad: false });
    }

    fetch('./README.md')
        .then((response) => {
            if (!response.ok) {
                throw new Error('Failed to load README');
            }
            return response.text();
        })
        .then((text) => {
            if (window.marked) {
                readmeZoomEl.innerHTML = window.marked.parse(text);
            } else {
                readmeZoomEl.textContent = text;
            }

            const mermaidBlocks = readmeZoomEl.querySelectorAll('pre code.language-mermaid');
            mermaidBlocks.forEach((block) => {
                const mermaidContainer = document.createElement('div');
                mermaidContainer.className = 'mermaid';
                mermaidContainer.textContent = block.textContent;
                const pre = block.parentElement;
                if (pre) {
                    pre.replaceWith(mermaidContainer);
                }
            });

            if (window.mermaid) {
                window.mermaid.run({ nodes: readmeZoomEl.querySelectorAll('.mermaid') });
            }
        })
        .catch(() => {
            readmeZoomEl.textContent = 'README not available.';
        });
}

loadReadme();

const readmeZoomEl = document.getElementById('readmeZoom');
const readmeContentEl = document.getElementById('readmeContent');

let zoomLevel = 1;
let baseDistance = null;
let baseZoom = 1;
const activePointers = new Map();

function applyZoom() {
    if (!readmeZoomEl) {
        return;
    }
    readmeZoomEl.style.transform = `scale(${zoomLevel})`;
}

function getDistance(a, b) {
    const dx = a.clientX - b.clientX;
    const dy = a.clientY - b.clientY;
    return Math.hypot(dx, dy);
}

if (readmeContentEl && readmeZoomEl) {
    readmeContentEl.addEventListener('pointerdown', (event) => {
        readmeContentEl.setPointerCapture(event.pointerId);
        activePointers.set(event.pointerId, event);
        if (activePointers.size === 2) {
            const [p1, p2] = Array.from(activePointers.values());
            baseDistance = getDistance(p1, p2);
            baseZoom = zoomLevel;
        }
    });

    readmeContentEl.addEventListener('pointermove', (event) => {
        if (!activePointers.has(event.pointerId)) {
            return;
        }
        activePointers.set(event.pointerId, event);
        if (activePointers.size === 2 && baseDistance) {
            const [p1, p2] = Array.from(activePointers.values());
            const currentDistance = getDistance(p1, p2);
            const scale = currentDistance / baseDistance;
            zoomLevel = Math.min(Math.max(baseZoom * scale, 0.6), 2.5);
            applyZoom();
        }
    });

    const endPointer = (event) => {
        if (activePointers.has(event.pointerId)) {
            activePointers.delete(event.pointerId);
        }
        if (activePointers.size < 2) {
            baseDistance = null;
        }
    };

    readmeContentEl.addEventListener('pointerup', endPointer);
    readmeContentEl.addEventListener('pointercancel', endPointer);
    readmeContentEl.addEventListener('pointerleave', endPointer);
}

function shortenURL() {
    const url = document.getElementById('urlInput').value;
    if (url === '') {
        alert('Please enter a URL');
        return;
    }

    const responseTimeEl = document.getElementById('responseTime');
    responseTimeEl.textContent = '';
    const resultEl = document.getElementById('shortenedResult');
    resultEl.classList.remove('show');

    fetch(`${host}/url/shorten`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ long_url: url })
    })
        .then(response => response.json())
        .then(data => {
            const shortUrl = data.data.short_url;
            const responseTimeMs = data.meta && typeof data.meta.responseTime === 'number'
                ? data.meta.responseTime
                : null;
            document.getElementById('shortenedLink').href = shortUrl;
            document.getElementById('shortenedLink').textContent = shortUrl;
            void resultEl.offsetHeight;
            resultEl.classList.add('show');
            responseTimeEl.textContent = responseTimeMs === null
                ? 'Response time: unavailable'
                : `Response time: ${responseTimeMs.toFixed(2)} ms`;
        })
        .catch(() => {
            responseTimeEl.textContent = 'Response time: unavailable (failed)';
        });
}

function copyToClipboard() {
    const shortenedURL = document.getElementById('shortenedLink').textContent;
    if (shortenedURL) {
        navigator.clipboard.writeText(shortenedURL).then(() => {
            alert("URL copied to clipboard!");
        }).catch((err) => {
            alert("Failed to copy URL: " + err);
        });
    }
}