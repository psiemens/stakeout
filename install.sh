#!/bin/sh

# Exit as soon as any command fails
set -e

BASE_URL="https://raw.githubusercontent.com/psiemens/stakeout/main"

# The architecture string, set by get_architecture
ARCH=""

# Get the architecture (CPU, OS) of the current system as a string.
# Only MacOS/x86_64 and Linux/x86_64 architectures are supported.
get_architecture() {
    _ostype="$(uname -s)"
    _cputype="$(uname -m)"
    _targetpath=""
    if [ "$_ostype" = Darwin ] && [ "$_cputype" = i386 ]; then
        if sysctl hw.optional.x86_64 | grep -q ': 1'; then
            _cputype=x86_64
        fi
    fi
    case "$_ostype" in
        Linux)
            _ostype=linux
            _targetpath=$HOME/.local/bin
            ;;
        Darwin)
            _ostype=darwin
            _targetpath=/usr/local/bin
            ;;
        *)
            echo "unrecognized OS type: $_ostype"
            return 1
            ;;
    esac
    case "$_cputype" in
        x86_64 | x86-64 | x64 | amd64)
            _cputype=x86_64
            ;;
        *)
            echo "unknown CPU type: $_cputype"
            return 1
            ;;
    esac
    _arch="${_cputype}-${_ostype}"
    ARCH="${_arch}"
    TARGET_PATH="${_targetpath}"
}

# Determine the system architecure, download the appropriate binary, and
# install it in `/usr/local/bin` on macOS and `~/.local/bin` on Linux
# with executable permission.
main() {

  get_architecture || exit 1

  echo "Downloading stakeout..."

  tmpfile=$(mktemp 2>/dev/null || mktemp -t stakeout)

  url="$BASE_URL/stakeout-$ARCH"
  curl --progress-bar "$url" -o $tmpfile

  # Ensure we don't receive a not found error as response.
  if grep -q "404: Not Found" $tmpfile
  then
    echo "\nCould not find compatible version for your operating system.\n"
    echo "Please open a GitHub issue with the following title:\n"
    echo "\"Failed to install for OS version $ARCH\""
    exit 1
  fi

  chmod +x $tmpfile

  [ -d $TARGET_PATH ] || mkdir -p $TARGET_PATH
  mv $tmpfile $TARGET_PATH/stakeout

  echo "Successfully installed Stakeout to $TARGET_PATH."
  echo "Make sure $TARGET_PATH is in your \$PATH environment variable.\n"
  echo "Example usage: \n"
  echo "stakeout 0xe467b9dd11fa00df"
}

main
