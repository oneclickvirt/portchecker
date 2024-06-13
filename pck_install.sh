#!/bin/bash
#From https://github.com/oneclickvirt/portchecker
#2024.05.23

rm -rf /usr/bin/pck
os=$(uname -s)
arch=$(uname -m)

case $os in
  Linux)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-linux-amd64
        ;;
      "i386" | "i686")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-linux-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-linux-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  Darwin)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-darwin-amd64
        ;;
      "i386" | "i686")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-darwin-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-darwin-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  FreeBSD)
    case $arch in
      amd64)
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-freebsd-amd64
        ;;
      "i386" | "i686")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-freebsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-freebsd-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  OpenBSD)
    case $arch in
      amd64)
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-openbsd-amd64
        ;;
      "i386" | "i686")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-openbsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O pck https://github.com/oneclickvirt/portchecker/releases/download/output/portchecker-openbsd-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported operating system: $os"
    exit 1
    ;;
esac

chmod 777 pck
if [ ! -f /usr/bin/pck ]; then
  mv pck /usr/bin/
fi
