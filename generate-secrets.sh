#!/bin/bash

# Production Secrets Generator
# Run this script to generate secure secrets for production deployment

echo "🔐 Generating Production Secrets..."
echo ""

# Generate JWT Secret (min 32 characters)
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET=$JWT_SECRET"

# Generate Internal Service Key (min 24 characters)
INTERNAL_SERVICE_KEY=$(openssl rand -base64 24)
echo "INTERNAL_SERVICE_KEY=$INTERNAL_SERVICE_KEY"

# Generate Database Passwords
DB_PASSWORD=$(openssl rand -base64 16)
echo "DB_PASSWORD=$DB_PASSWORD"

MONGODB_PASSWORD=$(openssl rand -base64 16)
echo "MONGODB_PASSWORD=$MONGODB_PASSWORD"

echo ""
echo "✅ Secrets generated successfully!"
echo ""
echo "📋 Add these to your production .env file:"
echo ""
cat << EOF
# PostgreSQL Database Configuration
DB_USER=posdigi_app_user
DB_PASSWORD=$DB_PASSWORD

# MongoDB Configuration
MONGODB_USERNAME=posdigi_app_user
MONGODB_PASSWORD=$MONGODB_PASSWORD

# JWT Configuration
JWT_SECRET=$JWT_SECRET

# Internal Service Communication
INTERNAL_SERVICE_KEY=$INTERNAL_SERVICE_KEY
EOF

echo ""
echo "⚠️  IMPORTANT: Store these secrets securely and never commit them to git!"
echo "💡 Save these secrets in a secure password manager."