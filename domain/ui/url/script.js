const host = "https://api.syamsul.online"
// const host = "http://0.0.0.0:5000"
function shortenURL() {
    const url = document.getElementById('urlInput').value;
    if (url === '') {
        alert('Please enter a URL');
        return;
    }

    var shortUrl
    fetch(`${host}/short?long_url=${url}`)  // Replace with your API URL
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json(); // Parse the JSON from the response
        })
        .then(data => {
            shortUrl = data.short_url
            // Update the displayed shortened URL
            document.getElementById('shortenedLink').href = shortUrl;
            document.getElementById('shortenedLink').textContent = shortUrl;
            document.getElementById('shortenedResult').style.display = 'block';
        })
        .catch(error => {
            console.error('There was a problem with the fetch operation:', error);
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