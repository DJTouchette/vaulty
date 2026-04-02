const express = require("express");
const app = express();
const port = process.env.PORT || 3000;

const apiKey = process.env.API_KEY || "";
const dbUrl = process.env.DB_URL || "";

// Show masked versions — never log raw secrets
function mask(s) {
  if (!s || s.length < 4) return "****";
  return s.slice(0, 4) + "****";
}

app.get("/", (req, res) => {
  res.json({
    status: "ok",
    api_key: mask(apiKey),
    db_url: mask(dbUrl),
  });
});

app.listen(port, () => {
  console.log(`Server running on port ${port}`);
  console.log(`API_KEY: ${mask(apiKey)}`);
  console.log(`DB_URL: ${mask(dbUrl)}`);
});
