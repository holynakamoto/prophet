#!/usr/bin/env node

const fs = require('fs-extra');
const path = require('path');
const express = require('express');
const serveStatic = require('serve-static');

const dir = path.resolve(__dirname);
const pkgs = path.join(dir, 'dist-pkg');
const port = process.env.PORT || 4500;

if (!fs.existsSync(pkgs)) {
  console.error('\n❌ Error: The dist-pkg directory doesn\'t exist.');
  console.error('   Run: npm run build\n');
  process.exit(1);
}

const app = express();

// Enable CORS for all routes (required for Rancher to load extensions)
app.use((req, res, next) => {
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', 'GET, OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type');
  res.header('Access-Control-Max-Age', '86400');
  
  if (req.method === 'OPTIONS') {
    return res.sendStatus(200);
  }
  next();
});

// Catalog endpoint
app.get('/', (req, res) => {
  const response = [];
  
  fs.readdirSync(pkgs).forEach((f) => {
    const pkgFile = path.join(pkgs, f, 'package.json');
    
    if (fs.existsSync(pkgFile)) {
      const rawdata = fs.readFileSync(pkgFile);
      const pkg = JSON.parse(rawdata);
      response.push(pkg);
    }
  });
  
  res.json(response);
});

// Serve static files
app.use(serveStatic(pkgs, {
  setHeaders: (res, filePath) => {
    // Ensure JavaScript files have correct content type
    if (filePath.endsWith('.js')) {
      res.setHeader('Content-Type', 'application/javascript; charset=UTF-8');
    }
  }
}));

app.listen(port, '0.0.0.0', () => {
  console.log('\n✅ Serving packages with CORS enabled:\n');
  console.log(`   http://127.0.0.1:${port}/k8sgpt-1.0.0/k8sgpt-1.0.0.umd.min.js`);
  console.log(`   http://0.0.0.0:${port}/k8sgpt-1.0.0/k8sgpt-1.0.0.umd.min.js\n`);
  console.log('📋 To load in Rancher:');
  console.log('   1. Go to Extensions → Developer Load');
  console.log(`   2. Enter: http://127.0.0.1:${port}/k8sgpt-1.0.0/k8sgpt-1.0.0.umd.min.js`);
  console.log('   3. Click Load\n');
});

