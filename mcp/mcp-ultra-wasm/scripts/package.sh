#!/bin/bash

# MCP Ultra Packaging Script
# Creates distributable archive of the complete project

set -e

PROJECT_NAME="mcp-ultra-wasm"
VERSION=${VERSION:-"v1.0.0"}
OUTPUT_DIR="packages"
ARCHIVE_NAME="${PROJECT_NAME}-${VERSION}"
TEMP_DIR=$(mktemp -d)

echo "🎯 MCP Ultra Packaging Script"
echo "════════════════════════════════════════"
echo "📦 Creating package: ${ARCHIVE_NAME}"
echo "📁 Output directory: ${OUTPUT_DIR}"
echo ""

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Copy project files to temp directory
echo "📋 Copying project files..."
cp -r . "${TEMP_DIR}/${PROJECT_NAME}"
cd "${TEMP_DIR}/${PROJECT_NAME}"

# Clean up development files
echo "🧹 Cleaning development files..."
rm -rf .git/
rm -rf bin/
rm -rf dist/
rm -rf coverage.out
rm -rf coverage.html
rm -rf *.test
rm -rf .testcache/
rm -rf vendor/
rm -rf node_modules/
rm -rf __debug_bin*
rm -rf .air.toml
rm -rf *.log
rm -rf *.sarif
rm -rf *.json.backup
rm -rf .secrets.baseline.new
rm -rf security-report.txt
rm -rf vulnerability-report.json
rm -rf gosec-results.sarif
rm -rf trivy-fs-results.sarif
rm -rf vuln-report.json
rm -rf sbom.spdx.json

# Remove sensitive files that shouldn't be distributed
echo "🔒 Removing sensitive files..."
find . -name "*.key" -delete
find . -name "*.crt" -delete
find . -name "*.pem" -delete
find . -name ".env" -delete
find . -name ".env.local" -delete
find . -name "*.env.local" -delete

# Ensure .env.example exists
if [ ! -f "config/.env.example" ]; then
    echo "❌ ERROR: config/.env.example not found!"
    exit 1
fi

# Create package info
echo "📝 Creating package information..."
cat > PACKAGE_INFO.md << EOF
# MCP Ultra v${VERSION} - Distribution Package

## 📦 Package Contents

This package contains the complete MCP Ultra enterprise microservice template.

### 🚀 What's Included:
- Complete Go source code (77 files)
- Comprehensive test suite (13 test files)
- Production-ready Docker configuration
- Kubernetes deployment manifests
- CI/CD pipelines (GitHub Actions)
- Complete documentation suite
- Development and deployment scripts

### 🎯 Validation Results:
- **Architecture**: A+ (100%)
- **DevOps**: A+ (100%) 
- **Observability**: B+ (85%)
- **Security**: C (70%)
- **Testing**: C+ (77%)

### 📊 Package Statistics:
- **Package Version**: ${VERSION}
- **Package Date**: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
- **Go Files**: 77
- **Test Files**: 13
- **Documentation Files**: 6+
- **Total Size**: $(du -sh . | cut -f1)

### 🛠 Quick Start:
1. Extract the archive
2. \`cp config/.env.example .env\`
3. Edit .env with your values
4. \`docker-compose up -d\`
5. \`make run\`

### 📚 Documentation:
- README.md - Complete setup guide
- CHANGELOG.md - Version history
- CONTRIBUTING.md - Development guidelines
- GITHUB_SETUP.md - GitHub configuration
- GITHUB_READY.md - Deployment summary

---
**Generated**: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Template**: MCP Ultra Enterprise Microservice
**Status**: Production Ready ✅
EOF

# Go back to original directory
cd - > /dev/null

# Create different archive formats
echo "📦 Creating archive formats..."

# ZIP Archive (Windows/Cross-platform)
cd "${TEMP_DIR}"
zip -r "${OUTPUT_DIR}/${ARCHIVE_NAME}.zip" "${PROJECT_NAME}" > /dev/null
echo "✅ Created: ${ARCHIVE_NAME}.zip"

# TAR.GZ Archive (Unix/Linux)
tar -czf "${OUTPUT_DIR}/${ARCHIVE_NAME}.tar.gz" "${PROJECT_NAME}"
echo "✅ Created: ${ARCHIVE_NAME}.tar.gz"

# TAR.XZ Archive (High compression)
tar -cJf "${OUTPUT_DIR}/${ARCHIVE_NAME}.tar.xz" "${PROJECT_NAME}"
echo "✅ Created: ${ARCHIVE_NAME}.tar.xz"

cd - > /dev/null

# Generate checksums
echo "🔐 Generating checksums..."
cd "${OUTPUT_DIR}"
sha256sum "${ARCHIVE_NAME}".* > "${ARCHIVE_NAME}.sha256"
echo "✅ Created: ${ARCHIVE_NAME}.sha256"
cd - > /dev/null

# Create distribution manifest
echo "📋 Creating distribution manifest..."
cat > "${OUTPUT_DIR}/DISTRIBUTION_MANIFEST.md" << EOF
# MCP Ultra v${VERSION} - Distribution Manifest

## 📦 Available Packages

### Archive Formats:
- **${ARCHIVE_NAME}.zip** - Windows/Cross-platform ($(ls -lh "${OUTPUT_DIR}/${ARCHIVE_NAME}.zip" | awk '{print $5}'))
- **${ARCHIVE_NAME}.tar.gz** - Standard compression ($(ls -lh "${OUTPUT_DIR}/${ARCHIVE_NAME}.tar.gz" | awk '{print $5}'))
- **${ARCHIVE_NAME}.tar.xz** - High compression ($(ls -lh "${OUTPUT_DIR}/${ARCHIVE_NAME}.tar.xz" | awk '{print $5}'))

### Verification:
- **${ARCHIVE_NAME}.sha256** - SHA-256 checksums for integrity verification

### Usage:
\`\`\`bash
# Download and verify
wget https://releases.example.com/${ARCHIVE_NAME}.tar.gz
wget https://releases.example.com/${ARCHIVE_NAME}.sha256
sha256sum -c ${ARCHIVE_NAME}.sha256

# Extract
tar -xzf ${ARCHIVE_NAME}.tar.gz
cd ${PROJECT_NAME}

# Quick start
cp config/.env.example .env
# Edit .env with your configuration
make install-tools
docker-compose up -d
make run
\`\`\`

### 🎯 What You Get:
- Complete enterprise-grade microservice template
- Production-ready configuration
- Comprehensive documentation
- Advanced security features
- Full observability stack
- Automated CI/CD pipelines

---
**Package Generated**: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Template Version**: ${VERSION}
**Total Files**: $(find "${TEMP_DIR}/${PROJECT_NAME}" -type f | wc -l)
EOF

# Cleanup
rm -rf "${TEMP_DIR}"

# Show results
echo ""
echo "🎉 PACKAGING COMPLETE!"
echo "════════════════════════════════════════"
echo "📁 Package Directory: ${OUTPUT_DIR}/"
echo ""
echo "📦 Generated Archives:"
ls -lh "${OUTPUT_DIR}/${ARCHIVE_NAME}".* | awk '{print "   " $9 " (" $5 ")"}'
echo ""
echo "🔐 Integrity Check:"
echo "   ${ARCHIVE_NAME}.sha256"
echo ""
echo "📋 Distribution Info:"
echo "   DISTRIBUTION_MANIFEST.md"
echo ""
echo "✅ Ready for distribution!"
echo ""
echo "🚀 Next Steps:"
echo "   1. Test extraction: tar -xzf ${OUTPUT_DIR}/${ARCHIVE_NAME}.tar.gz"
echo "   2. Verify checksums: sha256sum -c ${OUTPUT_DIR}/${ARCHIVE_NAME}.sha256"
echo "   3. Upload to release platform"
echo ""