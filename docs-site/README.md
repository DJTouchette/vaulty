# Vaulty Documentation Site

This is the Hugo-powered documentation site for Vaulty.

## Setup

```bash
cd docs-site

# Install the theme
git submodule add https://github.com/alex-shpak/hugo-book themes/hugo-book

# Run locally
hugo server -D

# Build for production
hugo --minify
```

The site will be available at `http://localhost:1313/vaulty/`.

## Structure

```
docs-site/
├── hugo.toml                          # Hugo configuration
├── content/
│   ├── _index.md                      # Homepage
│   └── docs/
│       ├── getting-started/           # Installation, first vault, first request
│       ├── guides/                    # MCP, templates, teams, keychain, frameworks, backends
│       ├── reference/                 # CLI, config, security model
│       └── architecture/             # System overview, redaction, audit
├── static/                            # Static assets (images, CSS)
└── themes/hugo-book/                  # Hugo Book theme (git submodule)
```

## Deployment

### GitHub Pages

Add this GitHub Actions workflow to `.github/workflows/docs.yml` to auto-deploy on push to `main`:

```yaml
name: Deploy Docs
on:
  push:
    branches: [main]
    paths: [docs-site/**]

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      pages: write
      id-token: write
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: true
      - uses: peaceiris/actions-hugo@v3
        with:
          hugo-version: latest
      - run: cd docs-site && hugo --minify
      - uses: actions/upload-pages-artifact@v3
        with:
          path: docs-site/public
      - id: deployment
        uses: actions/deploy-pages@v4
```
