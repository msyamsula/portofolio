# CV / Portfolio Website

A static HTML portfolio website for Muhammad Syamsul Arifin, designed with a newsprint/vintage aesthetic using Tailwind CSS.

## Structure

```
cv/
├── index.html          # Main landing page with overview, tech stack, and editorial
├── projects.html       # Projects showcase
├── experiences.html    # Work experience timeline
├── educations.html     # Educational background
├── contact.html        # Contact information
├── handle.js           # JavaScript for interactive features
│
├── files/              # Certificates, transcripts, and supporting documents
│   ├── transcript.jpg
│   ├── transkrip.jpg
│   ├── bachelor.jpg
│   ├── sarjana.jpg
│   ├── cloud-practitioner.pdf
│   ├── ielts.pdf
│   ├── best-paper.jpg
│   ├── ganesha-karsa.jpg
│   ├── gold-medalist.jpg
│   └── silver-medalist.jpg
│
├── front-page.png      # Professional portrait photo
├── itb.png             # ITB logo
└── cv.pdf              # PDF version of CV
```

## Tech Stack

- **HTML5** - Structure
- **Tailwind CSS** (via CDN) - Styling
- **Google Fonts** - Newsreader (serif), Inter (sans-serif), Material Symbols
- **Vanilla JavaScript** - Interactivity

## Design Features

- Newsprint/vintage paper aesthetic with subtle grain texture
- Dark mode support
- Responsive design (mobile-first approach)
- Smooth fade-in animations
- Serif display fonts with clean sans-serif body text

## Local Development

Simply open `index.html` in a browser, or use a local server:

```bash
# Python 3
python -m http.server 8000

# Node.js (with npx)
npx serve

# Then visit http://localhost:8000
```

## Deployment

This static site can be deployed to:

- **AWS S3 + CloudFront** (recommended for production with HTTPS)
- **GitHub Pages**
- **Netlify**
- **Vercel**
- Any static hosting service

### Quick S3 Deploy

```bash
# Sync to S3
aws s3 sync . s3://your-bucket-name --delete

# Set index document
aws s3 website s3://your-bucket-name --index-document index.html
```

### CloudFront + S3 (Recommended for Custom Domain)

1. Upload files to S3
2. Create CloudFront distribution
3. Add custom domain with ACM certificate
4. Configure Route 53

## Customization

To customize for your own use:

1. Replace personal information in HTML files
2. Update `front-page.png` with your photo
3. Modify `files/` contents with your certificates
4. Adjust Tailwind config in `<script id="tailwind-config">`
5. Update color scheme in CSS variables

## License

Personal use only.
