#!/usr/bin/env node

const { Binary } = require("binary-install");
const os = require("os");
const { version } = require("../package.json");

function getPlatform() {
  const type = os.type();
  const arch = os.arch();

  if (type === "Windows_NT") {
    return arch === "x64" ? "windows-amd64" : "windows-arm64";
  }
  if (type === "Linux") {
    return arch === "x64" ? "linux-amd64" : "linux-arm64";
  }
  if (type === "Darwin") {
    return arch === "x64" ? "darwin-amd64" : "darwin-arm64";
  }

  throw new Error(`Unsupported platform: ${type} ${arch}`);
}

function getBinary() {
  const platform = getPlatform();
  const url = `https://github.com/nehonix/watchtower/releases/download/v${version}/watchtower-${platform}.tar.gz`;
  const name = "watchtower";

  return new Binary(name, url, version);
}

const binary = getBinary();
binary.install();
