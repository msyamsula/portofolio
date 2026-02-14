require('dotenv').config();
const express = require('express');
const fs = require('fs/promises');
const path = require('path');
const emailjs = require('@emailjs/nodejs');

// EmailJS Configuration
const EMAILJS_PUBLIC_KEY = process.env.EMAILJS_PUBLIC_KEY;
const EMAILJS_PRIVATE_KEY = process.env.EMAILJS_PRIVATE_KEY || '';
const EMAILJS_SERVICE_ID = process.env.EMAILJS_SERVICE_ID;
const EMAILJS_TEMPLATE_ID = process.env.EMAILJS_TEMPLATE_ID;

const app = express();
const PORT = process.env.PORT || 2501;

// CORS headers for same-origin requests from subdirectory paths
app.use((req, res, next) => {
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');
    if (req.method === 'OPTIONS') {
        return res.sendStatus(200);
    }
    next();
});

app.use(express.json());

// Serve static files from current directory
app.use(express.static(__dirname));

// Serve static files from wedding directory
const weddingDir = path.join(__dirname, 'wedding');
app.use('/wedding', express.static(weddingDir));

// Wedding wishes API endpoints
const WEDDING_DATA_FILE = path.join(weddingDir, 'wishes.txt');
const LEGACY_WEDDING_DATA_FILE = path.join(weddingDir, 'whises.txt');

const readWishes = async () => {
    try {
        // Try new filename first
        let data = '';
        try {
            data = await fs.readFile(WEDDING_DATA_FILE, 'utf8');
        } catch (e) {
            // Fall back to old filename
            data = await fs.readFile(LEGACY_WEDDING_DATA_FILE, 'utf8');
        }
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
    const dataFile = WEDDING_DATA_FILE;
    try {
        existing = await fs.readFile(dataFile, 'utf8');
    } catch (error) {
        if (error.code !== 'ENOENT') {
            throw error;
        }
    }

    const nextContent = existing ? `${newLine}\n${existing}` : `${newLine}\n`;
    await fs.writeFile(dataFile, nextContent, 'utf8');
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

// EmailJS helper function for server-side
async function sendEmail(name, email, title, message) {
    try {
        // Send email using EmailJS Node.js SDK
        const response = await emailjs.send(
            EMAILJS_SERVICE_ID,
            EMAILJS_TEMPLATE_ID,
            {
                name: name,
                email: email,
                title: title,
                message: message,
            },
            {
                publicKey: EMAILJS_PUBLIC_KEY,
                privateKey: EMAILJS_PRIVATE_KEY,
            }
        );

        console.log('Email sent successfully!', response);
        return response;
    } catch (error) {
        console.error('EmailJS error:', error);
        throw error;
    }
}

// Contact form API endpoint
app.post('/api/contact', async (req, res) => {
    const name = String(req.body?.name || '').trim();
    const email = String(req.body?.email || '').trim();
    const title = String(req.body?.subject || '').trim();
    const message = String(req.body?.message || '').trim();

    // Validate inputs
    if (!name || !email || !message || !title) {
        return res.status(400).json({ error: 'All fields are required.' });
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
        return res.status(400).json({ error: 'Invalid email format.' });
    }

    try {
        console.log(name, email, title, message);
        
        // Send email using EmailJS
        await sendEmail(name, email, title, message);

        // Return success response
        res.json({
            success: true,
            message: 'Message sent successfully!'
        });
    } catch (error) {
        console.error('Error sending contact form:', error);
        res.status(500).json({ error: 'Failed to send message: ' + (error.message || error) });
    }
});

if (require.main === module) {
    app.listen(PORT, () => {
        console.log(`CV portfolio server running at http://localhost:${PORT}`);
        // console.log(`Wedding app available at http://localhost:${PORT}/wedding/index.html`);
    });
}

module.exports = app;
