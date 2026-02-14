const express = require('express');
const fs = require('fs/promises');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 2500;
const DATA_FILE = path.join(__dirname, 'whises.txt');

app.use(express.json());
app.use(express.static(__dirname));

const readWishes = async () => {
    try {
        const data = await fs.readFile(DATA_FILE, 'utf8');
        return data
            .split('\n')
            .map((line) => line.trim())
            .filter(Boolean)
            .map((line) => {
                const separatorIndex = line.indexOf(':');
                if (separatorIndex === -1) {
                    return { name: 'Anonymous', message: line };
                }
                return {
                    name: line.slice(0, separatorIndex).trim() || 'Anonymous',
                    message: line.slice(separatorIndex + 1).trim(),
                };
            });
    } catch (error) {
        if (error.code === 'ENOENT') {
            return [];
        }
        throw error;
    }
};

const prependWish = async (name, message) => {
    const sanitizedName = name.replace(/\r?\n/g, ' ').trim();
    const sanitizedMessage = message.replace(/\r?\n/g, ' ').trim();
    const newLine = `${sanitizedName}: ${sanitizedMessage}`;

    let existing = '';
    try {
        existing = await fs.readFile(DATA_FILE, 'utf8');
    } catch (error) {
        if (error.code !== 'ENOENT') {
            throw error;
        }
    }

    const nextContent = existing ? `${newLine}\n${existing}` : `${newLine}\n`;
    await fs.writeFile(DATA_FILE, nextContent, 'utf8');
};

app.get('/api/wishes', async (req, res) => {
    try {
        const wishes = await readWishes();
        res.json({ wishes });
    } catch (error) {
        res.status(500).json({ error: 'Failed to load wishes.' });
    }
});

app.post('/api/wishes', async (req, res) => {
    const name = String(req.body?.name || '').trim();
    const message = String(req.body?.message || '').trim();

    if (!name || !message) {
        return res.status(400).json({ error: 'Name and message are required.' });
    }

    try {
        await prependWish(name, message);
        const wishes = await readWishes();
        res.json({ wishes });
    } catch (error) {
        res.status(500).json({ error: 'Failed to save wish.' });
    }
});

if (require.main === module) {
    app.listen(PORT, () => {
        console.log(`Wedding wishes server running at http://localhost:${PORT}`);
    });
}

module.exports = app;
