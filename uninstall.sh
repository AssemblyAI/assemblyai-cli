#!/bin/sh

echo "Uninstalling AssemblyAI CLI..."
rm -rf "$HOME/.assemblyai-cli"

LINETOREMOVE="export PATH=\"$HOME/AssemblyAI:\$PATH\"" 
grep -v "$LINETOREMOVE" "$HOME/.zshrc" > "$HOME/.zshrc.tmp" && mv "$HOME/.zshrc.tmp" "$HOME/.zshrc"
grep -v "$LINETOREMOVE" "$HOME/.bashrc" > "$HOME/.bashrc.tmp" && mv "$HOME/.bashrc.tmp" "$HOME/.bashrc"

LINETOREMOVE="export PATH=\"$HOME/.assemblyai-cli:\$PATH\"" 
grep -v "$LINETOREMOVE" "$HOME/.zshrc" > "$HOME/.zshrc.tmp" && mv "$HOME/.zshrc.tmp" "$HOME/.zshrc"
grep -v "$LINETOREMOVE" "$HOME/.bashrc" > "$HOME/.bashrc.tmp" && mv "$HOME/.bashrc.tmp" "$HOME/.bashrc"

for path in $(echo "$PATH" | tr ":" "

"); do
  if [ -f "$path/assemblyai" ]; then
  sudo rm -f "$path/assemblyai"
  fi
done

echo "AssemblyAI CLI uninstalled."

if [ -f "$HOME/.bashrc" ]; then
	source "$HOME/.bashrc"
fi
if [ -f "$HOME/.zshrc" ]; then
	zsh
	source "$HOME/.zshrc"
fi