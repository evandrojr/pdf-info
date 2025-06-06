package main

import (
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// analyzePermissions analyzes PDF permissions and security settings
func (pa *PDFAnalyzer) analyzePermissions(ctx *model.Context, info *PDFInfo) {
	encDict, err := ctx.EncryptDict()
	if err != nil {
		// Se não conseguir obter o dicionário de criptografia, retornar
		return
	}

	// Verificar entradas U e O (senhas de usuário e proprietário)
	if _, foundU := encDict.Find("U"); foundU {
		info.UserPasswordSet = true
	}
	if _, foundO := encDict.Find("O"); foundO {
		info.OwnerPasswordSet = true
	}

	// Verificar permissões através do campo P
	if pVal := encDict.IntEntry("P"); pVal != nil {
		permissions := *pVal // Dereferencia pVal para obter o int
		
		// Analisar bits de permissão (PDF Reference)
		info.PrintAllowed = (permissions & 4) != 0
		info.ModifyAllowed = (permissions & 8) != 0
		info.CopyAllowed = (permissions & 16) != 0
		info.AddNotesAllowed = (permissions & 32) != 0
		info.FillFormsAllowed = (permissions & 256) != 0
		info.AccessibilityAllowed = (permissions & 512) != 0
		info.AssembleAllowed = (permissions & 1024) != 0
		info.PrintHighQualityAllowed = (permissions & 2048) != 0
	} else {
		// Valores padrão se P não for encontrado ou for nulo.
		// A especificação PDF pode ditar padrões restritivos se P estiver ausente em um PDF criptografado.
		// Para simplificar, definimos como true, mas isso pode não ser preciso para todos os casos.
		info.PrintAllowed = true
		info.ModifyAllowed = true
		info.CopyAllowed = true
		info.AddNotesAllowed = true
		info.FillFormsAllowed = true
		info.AccessibilityAllowed = true
		info.AssembleAllowed = true
		info.PrintHighQualityAllowed = true
	}
}
