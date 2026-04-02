# Vaulty + .env (Node.js example)

This example shows how to use Vaulty to manage `.env` secrets for a Node.js Express app.

## Walkthrough

### 1. Create a `.env` file

```bash
cp .env.example .env
# Edit .env with your real values
```

### 2. Import into Vaulty

```bash
vaulty import-env .env
```

This parses the `.env` file and stores each key-value pair in the encrypted vault. You'll see a warning if the file isn't in `.gitignore`.

### 3. View stored secrets (redacted)

```bash
vaulty export-env
```

Output:
```
API_KEY=****
DB_URL=****
PORT=****
```

### 4. Export with real values

```bash
vaulty export-env --reveal
```

### 5. Export to a file

```bash
vaulty export-env --reveal --out .env
```

### 6. Run the app with secrets injected

```bash
vaulty exec -- node app.js
```

Vaulty injects secrets as environment variables without writing them to disk.

## Key prefix

Use `--prefix` to namespace imported secrets:

```bash
vaulty import-env .env --prefix APP_
# Stores APP_API_KEY, APP_DB_URL, etc.
```
