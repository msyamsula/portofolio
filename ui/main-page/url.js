const host = "http://0.0.0.0:12000"
function shortenURL() {
    const url = document.getElementById('urlInput').value;
    if (url === '') {
        alert('Please enter a URL');
        return;
    }

    var shortUrl
    fetch(`${host}/short?long_url=${url}`)  // Replace with your API URL
        .then(response => {
            return response.json(); // Parse the JSON from the response
        })
        .then(data => {
            if (data.error !== "") {
                const error = new Error(data.error)
                throw error;
            }
            shortUrl = data.short_url
            // Update the displayed shortened URL
            document.getElementById('shortenedLink').href = shortUrl;
            document.getElementById('shortenedLink').textContent = shortUrl;
            document.getElementById('shortenedResult').style.display = 'block';
        })
        .catch(error => {
            alert(error);
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