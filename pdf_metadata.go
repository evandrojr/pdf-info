package main

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// analyzePDFCPU performs PDF analysis using the pdfcpu library
func (pa *PDFAnalyzer) analyzePDFCPU(filePath string, info *PDFInfo) error {
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return err
	}

	// Extract PDF metadata
	pa.extractMetadata(ctx, info)

	// Extract technical information
	pa.extractTechnicalInfo(ctx, info)

	// Extract structure information
	pa.extractStructureInfo(ctx, info)

	// Analyze security/permissions if encrypted
	if info.IsEncrypted && ctx.E != nil {
		pa.analyzePermissions(ctx, info)
	}

	// Analyze pages
	pa.analyzePages(ctx, info)

	// Analyze digital signatures
	pa.analyzeDigitalSignatures(filePath, ctx, info)

	return nil
}

// extractMetadata extracts PDF metadata from the Info dictionary
func (pa *PDFAnalyzer) extractMetadata(ctx *model.Context, info *PDFInfo) {
	if ctx.XRefTable != nil && ctx.XRefTable.Info != nil {
		infoObject, err := ctx.Dereference(*ctx.XRefTable.Info)
		if err != nil {
			fmt.Printf("Warning: could not dereference Info dictionary: %v\n", err)
		} else {
			if actualInfoDict, ok := infoObject.(types.Dict); ok {
				info.Title = getStringFromDict(actualInfoDict, "Title")
				info.Author = getStringFromDict(actualInfoDict, "Author")
				info.Subject = getStringFromDict(actualInfoDict, "Subject")
				info.Keywords = getStringFromDict(actualInfoDict, "Keywords")
				info.Creator = getStringFromDict(actualInfoDict, "Creator")
				info.Producer = getStringFromDict(actualInfoDict, "Producer")
				info.CreationDate = getStringFromDict(actualInfoDict, "CreationDate")
				info.ModDate = getStringFromDict(actualInfoDict, "ModDate")
			} else {
				fmt.Printf("Warning: Info object is not a dictionary, but rather type %T\n", infoObject)
			}
		}
	}
}

// extractTechnicalInfo extracts technical PDF information
func (pa *PDFAnalyzer) extractTechnicalInfo(ctx *model.Context, info *PDFInfo) {
	if ctx.HeaderVersion != nil {
		// HeaderVersion contém apenas a parte decimal (3 para 1.3, 4 para 1.4, etc.)
		// Adiciona o "1." na frente para formar a versão completa
		info.PDFVersion = fmt.Sprintf("1.%d", *ctx.HeaderVersion)
	}
	info.PageCount = ctx.PageCount
	info.IsEncrypted = ctx.E != nil

	// Verificar linearização através de propriedades do contexto
	info.IsLinearized = ctx.LinearizationObjs != nil
}

// extractStructureInfo extracts structural information from the PDF
func (pa *PDFAnalyzer) extractStructureInfo(ctx *model.Context, info *PDFInfo) {
	if ctx.RootDict != nil {
		// Verificar se tem formulários
		if entry := ctx.RootDict.DictEntry("AcroForm"); entry != nil {
			info.HasForms = true
		}

		// Verificar JavaScript
		if namesDict := ctx.RootDict.DictEntry("Names"); namesDict != nil {
			if jsEntry := namesDict.DictEntry("JavaScript"); jsEntry != nil {
				info.HasJavaScript = true
			}
			// Verificar anexos
			if efEntry := namesDict.DictEntry("EmbeddedFiles"); efEntry != nil {
				info.HasAttachments = true
				pa.extractAttachments(ctx, info)
			}
		}

		// Verificar marcadores
		if outlinesEntry := ctx.RootDict.DictEntry("Outlines"); outlinesEntry != nil {
			info.HasBookmarks = true
			pa.extractBookmarks(ctx, info)
		}

		// Verificar se é tagged (acessível)
		if markInfoDict := ctx.RootDict.DictEntry("MarkInfo"); markInfoDict != nil {
			if markedVal := markInfoDict.BooleanEntry("Marked"); markedVal != nil && *markedVal {
				info.IsTagged = true
			}
		}
	}
}
