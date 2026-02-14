const express = require('express');

const app = express();
const PORT = process.env.PORT || 2501;

// Serve static files from current directory
app.use(express.static(__dirname));

if (require.main === module) {
    app.listen(PORT, () => {
        console.log(`CV portfolio server running at http://localhost:${PORT}`);
    });
}

module.exports = app;
