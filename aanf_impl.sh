#!/bin/bash

# Nome do diretório onde o repositório será clonado
TARGET_DIR="nf_aanf_ausf"
REPO_URL="https://github.com/gustavojardim/ausf"

# Verifica se o diretório já existe
if [ -d "$TARGET_DIR" ]; then
  echo "O diretório '$TARGET_DIR' já existe. Atualizando com git pull..."
  cd "$TARGET_DIR" || { echo "Erro ao acessar o diretório '$TARGET_DIR'."; exit 1; }
  git pull
else
  echo "O diretório '$TARGET_DIR' não existe. Clonando o repositório..."
  git clone "$REPO_URL" "$TARGET_DIR"
fi

echo "Operação concluída."

