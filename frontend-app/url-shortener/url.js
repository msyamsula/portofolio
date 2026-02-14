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