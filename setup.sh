#!/bin/sh

# Function to get the latest release download URL for the specified OS and architecture
get_latest_release_url() {
  repo="baalimago/repeater"
  os="$1"
  arch="$2"

  # Fetch the latest release data from GitHub API
  release_data=$(curl -s "https://api.github.com/repos/$repo/releases/latest")

  # Extract the asset URL for the specified OS and architecture
  download_url=$(echo "$release_data" | grep "browser_download_url" | grep "$os" | grep "$arch" | cut -d '"' -f 4)

  echo "$download_url"
}

# Detect the OS
case "$(uname)" in
  Linux*)
    os="linux"
    ;;
  Darwin*)
    os="darwin"
    ;;
  *)
    echo "Unsupported OS: $(uname)"
    exit 1
    ;;
esac

# Detect the architecture
arch=$(uname -m)
case "$arch" in
  x86_64)
    arch="amd64"
    ;;
  armv7*)
    arch="arm"
    ;;
  aarch64|arm64)
    arch="arm64"
    ;;
  i?86)
    arch="386"
    ;;
  *)
    echo "Unsupported architecture: $arch"
    exit 1
    ;;
esac

printf "detected os: '%s', arch: '%s'\n" "$os" "$arch"

# Get the download URL for the latest release
printf "finding asset url..."
download_url=$(get_latest_release_url "$os" "$arch")
printf "OK!\n"

# Download the binary
tmp_file=$(mktemp)

printf "downloading binary..."
if ! curl -s -L -o "$tmp_file" "$download_url"; then
  echo
  echo "Failed to download the binary."
  exit 1
fi
printf "OK!\n"

printf "setting file executable file permissions..."
# Make the binary executable

if ! chmod +x "$tmp_file"; then
  echo
  echo "Failed to make the binary executable. Try running the script with sudo."
  exit 1
fi
printf "OK!\n"

# Move the binary to /usr/local/bin and handle permission errors
if ! mv "$tmp_file" /usr/local/bin/repeater; then
  echo "Failed to move the binary to /usr/local/bin/repeater, see error above. Try running the script with sudo, or run 'mv $tmp_file <desired-position>'."
  exit 1
fi

echo "repeater installed successfully in /usr/local/bin, try it out with 'repeater -h'"
