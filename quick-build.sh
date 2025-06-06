#!/bin/bash

# Script para compilar releases para as principais plataformas
set -e

APP_NAME="pdf-info"
SOURCE_FILE="pdf-info.go"
RELEASES_DIR="releases"

# Cores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}Compilando PDF Analysis Tool para múltiplas plataformas...${NC}"
echo

# Criar diretório de releases se não existir
mkdir -p "$RELEASES_DIR"

# Função para compilar e comprimir
build_and_compress() {
    local goos=$1
    local goarch=$2
    local output_name=$3
    
    echo -e "${GREEN}Compilando ${goos}/${goarch}...${NC}"
    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w -extldflags '-static'" -a -installsuffix cgo -o "$output_name" "$SOURCE_FILE"
    
    echo -e "${BLUE}Comprimindo ${output_name}...${NC}"
    gzip -9 "$output_name"
    echo "✓ Criado: ${output_name}.gz (arquivo não compactado removido)"
}

# Compilações principais
build_and_compress "windows" "amd64" "$RELEASES_DIR/${APP_NAME}-windows-amd64.exe"
# build_and_compress "windows" "386" "$RELEASES_DIR/${APP_NAME}-windows-386.exe"

build_and_compress "linux" "amd64" "$RELEASES_DIR/${APP_NAME}-linux-amd64"
# build_and_compress "linux" "386" "$RELEASES_DIR/${APP_NAME}-linux-386"
build_and_compress "linux" "arm64" "$RELEASES_DIR/${APP_NAME}-linux-arm64"
build_and_compress "linux" "arm" "$RELEASES_DIR/${APP_NAME}-linux-arm"

build_and_compress "darwin" "amd64" "$RELEASES_DIR/${APP_NAME}-darwin-amd64"
build_and_compress "darwin" "arm64" "$RELEASES_DIR/${APP_NAME}-darwin-arm64"

build_and_compress "freebsd" "amd64" "$RELEASES_DIR/${APP_NAME}-freebsd-amd64"

build_and_compress "openbsd" "amd64" "$RELEASES_DIR/${APP_NAME}-openbsd-amd64"

echo
echo -e "${BLUE}Compilações concluídas!${NC}"

# Mostrar arquivos criados
echo
echo -e "${GREEN}Binários comprimidos criados:${NC}"
ls -lh "$RELEASES_DIR"/${APP_NAME}-*.gz

# Criar README geral
cat > "$RELEASES_DIR/README.md" << 'EOF'
# PDF Analysis Tool - Releases

Binários compilados para múltiplas plataformas.

## Downloads

### Windows
- `pdf-info-windows-amd64.exe.gz` - Windows 64-bit (comprimido)
- `pdf-info-windows-386.exe.gz` - Windows 32-bit (comprimido)

### Linux  
- `pdf-info-linux-amd64.gz` - Linux 64-bit (comprimido)
- `pdf-info-linux-386.gz` - Linux 32-bit (comprimido)
- `pdf-info-linux-arm64.gz` - Linux ARM64 (comprimido)
- `pdf-info-linux-arm.gz` - Linux ARM (comprimido)

### macOS
- `pdf-info-darwin-amd64.gz` - macOS Intel (comprimido)
- `pdf-info-darwin-arm64.gz` - macOS Apple Silicon (comprimido)

### BSD
- `pdf-info-freebsd-amd64.gz` - FreeBSD 64-bit (comprimido)
- `pdf-info-openbsd-amd64.gz` - OpenBSD 64-bit (comprimido)

## Como usar

```bash
# Primeiro descomprimir
gunzip pdf-info-linux-amd64.gz

# Depois executar
# Linux/macOS/BSD
./pdf-info-linux-amd64 documento.pdf

# Windows (PowerShell)
Expand-Archive pdf-info-windows-amd64.exe.gz .
pdf-info-windows-amd64.exe documento.pdf
```

## Funcionalidades

- ✅ Análise de metadados PDF
- ✅ Verificação de assinaturas digitais
- ✅ Informações sobre criptografia
- ✅ Análise de segurança
- ✅ Detecção de formulários
- ✅ Informações do autor e criação

## Verificação

Para verificar a integridade dos arquivos, use:

```bash
# Primeiro descomprimir
gunzip pdf-info-linux-amd64.gz

# Verificar se o arquivo não está corrompido
file pdf-info-linux-amd64

# Testar execução
./pdf-info-linux-amd64 -help
```
EOF

echo
echo -e "${GREEN}README criado: ${RELEASES_DIR}/README.md${NC}"
echo -e "${BLUE}Todas as releases estão prontas na pasta: ${RELEASES_DIR}${NC}"
