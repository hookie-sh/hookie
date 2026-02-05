import express from "express";

const app = express();
const PORT = process.env.PORT || 3001;

// Middleware to parse JSON bodies
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Catch-all route handler for all HTTP methods
app.use((req, res) => {
  const requestInfo = {
    method: req.method,
    path: req.path,
    url: req.url,
    query: req.query,
    headers: req.headers,
    body: req.body,
    timestamp: new Date().toISOString(),
  };

  console.log("Incoming request:", JSON.stringify(requestInfo, null, 2));

  res.json(requestInfo);
});

app.listen(PORT, () => {
  console.log(`Express playground server running on http://localhost:${PORT}`);
  console.log("All endpoints will display incoming request information");
});
