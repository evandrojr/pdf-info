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
