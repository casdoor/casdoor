#!/bin/bash
set -e

echo "🔧 Overriding Casdoor login and signup pages..."

pwd
cp ../override/LoginPage.js ./src/auth/LoginPage.js
cp ../override/SignupPage.js ./src/auth/SignupPage.js

echo "✅ Override complete."
