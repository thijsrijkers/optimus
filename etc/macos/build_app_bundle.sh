#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

APP_NAME="${APP_NAME:-Optimus}"
BUNDLE_ID="${BUNDLE_ID:-dev.optimus.app}"
VERSION="${VERSION:-0.1.0}"
ICON_PNG="${ICON_PNG:-${ROOT_DIR}/etc/app_icon.png}"
OUT_DIR="${OUT_DIR:-${ROOT_DIR}/dist/macos}"

APP_DIR="${OUT_DIR}/${APP_NAME}.app"
CONTENTS_DIR="${APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
EXECUTABLE="${APP_NAME}"

if ! command -v iconutil >/dev/null 2>&1; then
  echo "error: iconutil is required (macOS)."
  exit 1
fi

if ! command -v sips >/dev/null 2>&1; then
  echo "error: sips is required (macOS)."
  exit 1
fi

if [[ ! -f "${ICON_PNG}" ]]; then
  echo "error: icon PNG not found at ${ICON_PNG}"
  exit 1
fi

rm -rf "${APP_DIR}"
mkdir -p "${MACOS_DIR}" "${RESOURCES_DIR}"

echo "Building ${APP_NAME}.app"
go build -o "${MACOS_DIR}/${EXECUTABLE}" "${ROOT_DIR}"

ICONSET_DIR="${OUT_DIR}/AppIcon.iconset"
rm -rf "${ICONSET_DIR}"
mkdir -p "${ICONSET_DIR}"

for size in 16 32 128 256 512; do
  sips -z "${size}" "${size}" "${ICON_PNG}" --out "${ICONSET_DIR}/icon_${size}x${size}.png" >/dev/null
  scale2=$((size * 2))
  sips -z "${scale2}" "${scale2}" "${ICON_PNG}" --out "${ICONSET_DIR}/icon_${size}x${size}@2x.png" >/dev/null
done

iconutil -c icns "${ICONSET_DIR}" -o "${RESOURCES_DIR}/AppIcon.icns"
rm -rf "${ICONSET_DIR}"

cat >"${CONTENTS_DIR}/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key>
  <string>en</string>
  <key>CFBundleExecutable</key>
  <string>${EXECUTABLE}</string>
  <key>CFBundleIconFile</key>
  <string>AppIcon</string>
  <key>CFBundleIdentifier</key>
  <string>${BUNDLE_ID}</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundleName</key>
  <string>${APP_NAME}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>${VERSION}</string>
  <key>CFBundleVersion</key>
  <string>${VERSION}</string>
  <key>LSMinimumSystemVersion</key>
  <string>12.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
EOF

echo "Done: ${APP_DIR}"
echo "Open it with: open \"${APP_DIR}\""
echo "Install system-wide with: cp -R \"${APP_DIR}\" /Applications/"
