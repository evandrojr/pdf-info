package main

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type PDFInfo struct {
	// Informações básicas do arquivo
	FileName     string
	FilePath     string
	FileSize     int64
	FileSizeHuman string
	LastModified time.Time
	MD5Hash      string
	SHA256Hash   string

	// Informações do documento PDF
	Title        string
	Author       string
	Subject      string
	Keywords     string
	Creator      string
	Producer     string
	CreationDate string
	ModDate      string

	// Informações técnicas
	PDFVersion    string
	PageCount     int
	IsEncrypted   bool
	IsLinearized  bool
	IsTagged      bool
	HasBookmarks  bool
	HasAttachments bool
	HasForms      bool
	HasJavaScript bool
	HasAnnotations bool

	// Informações de segurança
	UserPasswordSet  bool
	OwnerPasswordSet bool
	PrintAllowed     bool
	ModifyAllowed    bool
	CopyAllowed      bool
	AddNotesAllowed  bool
	FillFormsAllowed bool
	AccessibilityAllowed bool
	AssembleAllowed  bool
	PrintHighQualityAllowed bool

	// Informações de assinatura digital
	HasDigitalSignatures bool
	SignatureCount       int
	Signatures          []DigitalSignatureInfo

	// Informações das páginas
	Pages []PageInfo

	// Informações de conteúdo
	TotalTextLength int
	FontsUsed       []string
	ImagesCount     int
	
	// Informações extras
	Bookmarks    []BookmarkInfo
	Attachments  []AttachmentInfo
	Annotations  []AnnotationInfo
}

type PageInfo struct {
	Number     int
	Width      float64
	Height     float64
	Rotation   int
	TextLength int
	ImageCount int
}

type BookmarkInfo struct {
	Title string
	Level int
	Page  int
}

type AttachmentInfo struct {
	Name string
	Size int64
	Type string
}

type AnnotationInfo struct {
	Type    string
	Page    int
	Content string
}

type DigitalSignatureInfo struct {
	Type          string
	SubFilter     string
	SignerName    string
	SigningTime   string
	Location      string
	Reason        string
	ContactInfo   string
	FieldName     string
	IsValid       bool
	IsCertified   bool
	Status        string
	ValidationErrors []string
	
	// Timestamp information
	HasTimestamp     bool
	TimestampType    string
	TimestampTime    string
	TimestampAuthority string
	TimestampStatus  string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <pdf_path>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]
	
	analyzer := &PDFAnalyzer{}
	info, err := analyzer.AnalyzePDF(pdfPath)
	if err != nil {
		log.Fatalf("Error analyzing PDF: %v", err)
	}

	analyzer.PrintReport(info)
}

type PDFAnalyzer struct{}

func (pa *PDFAnalyzer) AnalyzePDF(filePath string) (*PDFInfo, error) {
	info := &PDFInfo{}
	
	// Basic file information
	if err := pa.getFileInfo(filePath, info); err != nil {
		return nil, fmt.Errorf("error getting file information: %v", err)
	}

	// Analysis using pdfcpu
	if err := pa.analyzePDFCPU(filePath, info); err != nil {
		fmt.Printf("Warning: error in pdfcpu analysis: %v\n", err)
	}

	// Analysis using ledongthuc/pdf
	if err := pa.analyzeLedongthuc(filePath, info); err != nil {
		fmt.Printf("Warning: error in ledongthuc analysis: %v\n", err)
	}

	return info, nil
}

func (pa *PDFAnalyzer) getFileInfo(filePath string, info *PDFInfo) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	info.FileName = filepath.Base(filePath)
	info.FilePath = filePath
	info.FileSize = stat.Size()
	info.FileSizeHuman = formatFileSize(stat.Size())
	info.LastModified = stat.ModTime()

	// Calcular hashes
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// MD5
	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, file); err != nil {
		return err
	}
	info.MD5Hash = fmt.Sprintf("%x", md5Hash.Sum(nil))

	// SHA256
	file.Seek(0, 0)
	sha256Hash := sha256.New()
	if _, err := io.Copy(sha256Hash, file); err != nil {
		return err
	}
	info.SHA256Hash = fmt.Sprintf("%x", sha256Hash.Sum(nil))

	return nil
}

func (pa *PDFAnalyzer) analyzePDFCPU(filePath string, info *PDFInfo) error {
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return err
	}

	// Informações básicas
	if ctx.XRefTable != nil && ctx.XRefTable.Info != nil {
		infoObject, err := ctx.Dereference(*ctx.XRefTable.Info)
		if err != nil {
			fmt.Printf("Warning: could not dereference Info dictionary: %v\\n", err)
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
				fmt.Printf("Warning: Info object is not a dictionary, but rather type %T\\n", infoObject)
			}
		}
	}

	// Informações técnicas
	if ctx.HeaderVersion != nil {
		// HeaderVersion contém apenas a parte decimal (3 para 1.3, 4 para 1.4, etc.)
		// Adiciona o "1." na frente para formar a versão completa
		info.PDFVersion = fmt.Sprintf("1.%d", *ctx.HeaderVersion)
	}
	info.PageCount = ctx.PageCount
	info.IsEncrypted = ctx.E != nil

	// Verificar linearização através de propriedades do contexto
	info.IsLinearized = ctx.LinearizationObjs != nil

	// Verificar se tem formulários
	if ctx.RootDict != nil {
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

	// Informações de segurança se criptografado
	if info.IsEncrypted && ctx.E != nil {
		pa.analyzePermissions(ctx, info)
	}

	// Analisar páginas
	pa.analyzePages(ctx, info)

	// Analisar assinaturas digitais
	pa.analyzeDigitalSignatures(filePath, ctx, info)

	return nil
}

func (pa *PDFAnalyzer) analyzeLedongthuc(filePath string, info *PDFInfo) error {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Extrair texto de todas as páginas
	totalTextLength := 0
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		
		textLen := len(strings.TrimSpace(text))
		totalTextLength += textLen
		
		// Atualizar informação da página se ela existir
		if i-1 < len(info.Pages) {
			info.Pages[i-1].TextLength = textLen
		}
	}
	
	info.TotalTextLength = totalTextLength
	return nil
}

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

func (pa *PDFAnalyzer) analyzePages(ctx *model.Context, info *PDFInfo) {
	info.Pages = make([]PageInfo, ctx.PageCount)
	
	for i := 1; i <= ctx.PageCount; i++ {
		pageInfo := PageInfo{
			Number: i,
		}

		// Obter informações da página
		pageDict, _, _, err := ctx.PageDict(i, false)
		if err == nil && pageDict != nil {
			// MediaBox para dimensões
			if mediaBox := pageDict.ArrayEntry("MediaBox"); mediaBox != nil && len(mediaBox) >= 4 {
				if width, ok := mediaBox[2].(types.Float); ok {
					pageInfo.Width = float64(width)
				}
				if height, ok := mediaBox[3].(types.Float); ok {
					pageInfo.Height = float64(height)
				}
			}

			// Rotação
			if rotate := pageDict.IntEntry("Rotate"); rotate != nil {
				pageInfo.Rotation = *rotate
			}

			// Verificar se há anotações
			if annots := pageDict.ArrayEntry("Annots"); annots != nil && len(annots) > 0 {
				info.HasAnnotations = true
			}
		}

		info.Pages[i-1] = pageInfo
	}
}

func (pa *PDFAnalyzer) extractBookmarks(ctx *model.Context, info *PDFInfo) {
	// Implementação simplificada para bookmarks
	info.Bookmarks = []BookmarkInfo{
		{Title: "Marcadores detectados no documento", Level: 1, Page: 1},
	}
}

func (pa *PDFAnalyzer) extractAttachments(ctx *model.Context, info *PDFInfo) {
	// Implementação simplificada para anexos
	info.Attachments = []AttachmentInfo{
		{Name: "Anexos detectados no documento", Size: 0, Type: "Detectado"},
	}
}

func (pa *PDFAnalyzer) analyzeDigitalSignatures(filePath string, ctx *model.Context, info *PDFInfo) {
	// First, try to detect signature fields directly from the PDF structure
	// This works even for encrypted PDFs in many cases
	hasSignatureFields := pa.detectSignatureFields(ctx, info)
	
	// If structural analysis fails, try raw byte analysis
	if !hasSignatureFields {
		hasRawSignatures, rawCount, err := pa.detectSignaturesByteAnalysis(filePath)
		if err != nil {
			// Silent error - continue with no signatures detected
		} else if hasRawSignatures {
			info.HasDigitalSignatures = true
			info.SignatureCount = rawCount
			hasSignatureFields = true
		}
	}
	
	// Try to validate signatures using pdfcpu (this may fail for encrypted PDFs)
	results, err := api.ValidateSignatures(filePath, true, nil) // all=true
	if err != nil {
		fmt.Printf("Warning: error validating signatures: %v", err)
		// If validation fails but we detected signature fields, still report them
		if hasSignatureFields {
			info.HasDigitalSignatures = true
			// Keep the signature count from field detection
			return
		}
		info.HasDigitalSignatures = false
		info.SignatureCount = 0
		return
	}

	if len(results) == 0 {
		// If no validation results but we found signature fields, report the fields
		if hasSignatureFields {
			info.HasDigitalSignatures = true
			// Keep the signature count from field detection
			return
		}
		info.HasDigitalSignatures = false
		info.SignatureCount = 0
		return
	}

	// We have successful validation results
	info.HasDigitalSignatures = true
	info.SignatureCount = len(results)
	info.Signatures = make([]DigitalSignatureInfo, 0, len(results))

	// Process each validation result
	for _, result := range results {
		sigInfo := DigitalSignatureInfo{
			FieldName:   result.Details.FieldName,
			SubFilter:   result.Details.SubFilter,
			SignerName:  result.Details.SignerName,
			SigningTime: formatTime(result.Details.SigningTime),
			Location:    result.Details.Location,
			Reason:      result.Details.Reason,
			ContactInfo: result.Details.ContactInfo,
			IsCertified: result.Certified(),
			IsValid:     result.Status == 1, // SignatureStatusValid
		}

		// Determine signature status
		switch result.Status {
		case 1: // model.SignatureStatusValid
			sigInfo.Status = "Valid"
		case 2: // model.SignatureStatusInvalid
			sigInfo.Status = "Invalid"
		default: // model.SignatureStatusUnknown or others
			sigInfo.Status = "Unknown"
		}

		// Add validation errors if any exist
		if len(result.Problems) > 0 {
			sigInfo.ValidationErrors = result.Problems
		}

		// Determine signature type
		if result.Certified() {
			sigInfo.Type = "Certified"
		} else {
			sigInfo.Type = "Approval"
		}

		// Analyze timestamp information
		pa.analyzeTimestamp(filePath, &sigInfo)

		info.Signatures = append(info.Signatures, sigInfo)
	}
}

// analyzeTimestamp detects and analyzes timestamp information in signatures
func (pa *PDFAnalyzer) analyzeTimestamp(filePath string, sigInfo *DigitalSignatureInfo) {
	// Initialize timestamp fields
	sigInfo.HasTimestamp = false
	sigInfo.TimestampType = ""
	sigInfo.TimestampTime = ""
	sigInfo.TimestampAuthority = ""
	sigInfo.TimestampStatus = "None"

	// Try to detect timestamp by analyzing raw PDF content
	hasTimestamp, timestampInfo := pa.detectTimestampByteAnalysis(filePath)
	if hasTimestamp {
		sigInfo.HasTimestamp = true
		sigInfo.TimestampType = timestampInfo["type"]
		sigInfo.TimestampTime = timestampInfo["time"]
		sigInfo.TimestampAuthority = timestampInfo["authority"]
		sigInfo.TimestampStatus = "Present"
	}
}

// detectTimestampByteAnalysis performs raw byte analysis for timestamp detection
func (pa *PDFAnalyzer) detectTimestampByteAnalysis(filePath string) (bool, map[string]string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, nil
	}

	content := string(data)
	timestampInfo := make(map[string]string)
	
	// Look for timestamp-related patterns
	timestampPatterns := []string{
		"/SubFilter/ETSI.RFC3161",
		"/SubFilter/adbe.pkcs7.detached",
		"/M(D:",  // Timestamp marker
		"/TS",    // Timestamp token
		"1.2.840.113549.1.9.16.1.4",  // RFC3161 timestamp OID
		"TimeStampToken",
		"TSATimeStamp",
		"timestampToken",
		"/Type/TSA",
		"Serpro",  // Serpro timestamp authority
		"Assinador Serpro",  // Serpro signer/timestamp
		"ICP-Brasil",  // ICP-Brasil timestamp
		"AC Timestamping",  // Certificate Authority Timestamp
		"Carimbo",  // Portuguese for timestamp
		"D:20",  // Date timestamp format
		"/ByteRange",  // Signature byte range (often indicates timestamps)
	}

	hasTimestamp := false
	for _, pattern := range timestampPatterns {
		if strings.Contains(content, pattern) {
			hasTimestamp = true
			
			// Determine timestamp type based on pattern
			if strings.Contains(pattern, "RFC3161") {
				timestampInfo["type"] = "RFC3161"
			} else if strings.Contains(pattern, "adbe.pkcs7") {
				timestampInfo["type"] = "PKCS#7"
			} else if strings.Contains(pattern, "TSA") {
				timestampInfo["type"] = "TSA"
			} else if strings.Contains(pattern, "Serpro") || strings.Contains(pattern, "Assinador Serpro") {
				timestampInfo["type"] = "Serpro TSA"
			} else if strings.Contains(pattern, "ICP-Brasil") {
				timestampInfo["type"] = "ICP-Brasil"
			} else if strings.Contains(pattern, "ByteRange") {
				timestampInfo["type"] = "Embedded"
			} else {
				timestampInfo["type"] = "Standard"
			}
			break
		}
	}

	// Try to extract timestamp time if found
	if hasTimestamp {
		// Look for timestamp time patterns
		timePatterns := []string{
			"D:202", // Year 2020+
			"D:201", // Year 2010+
		}
		
		for _, timePattern := range timePatterns {
			if idx := strings.Index(content, timePattern); idx != -1 {
				// Extract timestamp date
				endIdx := idx + 20
				if endIdx > len(content) {
					endIdx = len(content)
				}
				timestampStr := content[idx:endIdx]
				
				// Clean up and format the timestamp string
				if len(timestampStr) >= 15 {
					// Format: D:YYYYMMDDHHmmSS -> YYYY-MM-DD HH:mm:SS
					rawTime := timestampStr
					if strings.HasPrefix(rawTime, "D:") && len(rawTime) >= 15 {
						year := rawTime[2:6]
						month := rawTime[6:8]
						day := rawTime[8:10]
						hour := rawTime[10:12]
						minute := rawTime[12:14]
						second := "00"
						if len(rawTime) >= 16 {
							second = rawTime[14:16]
						}
						timestampInfo["time"] = fmt.Sprintf("%s-%s-%s %s:%s:%s", year, month, day, hour, minute, second)
					} else {
						timestampInfo["time"] = rawTime[:15]
					}
				} else {
					timestampInfo["time"] = timestampStr
				}
				break
			}
		}
		
		// Try to detect timestamp authority
		authorityPatterns := []string{
			"CN=",
			"O=",
			"TSA",
			"TimeStamp",
			"Serpro",
			"Assinador Serpro",
			"ICP-Brasil",
			"AC Timestamping",
			"Autoridade Certificadora",
			"Certificate Authority",
		}
		
		for _, authPattern := range authorityPatterns {
			if idx := strings.Index(content, authPattern); idx != -1 {
				// Extract authority name
				endIdx := idx + 80
				if endIdx > len(content) {
					endIdx = len(content)
				}
				authStr := content[idx:endIdx]
				
				// Clean up authority string
				if newlineIdx := strings.Index(authStr, "\n"); newlineIdx != -1 {
					authStr = authStr[:newlineIdx]
				}
				if nullIdx := strings.Index(authStr, "\x00"); nullIdx != -1 {
					authStr = authStr[:nullIdx]
				}
				
				// Specific cleanup for known authorities
				if strings.Contains(authStr, "Serpro") || strings.Contains(authStr, "Assinador Serpro") {
					timestampInfo["authority"] = "Serpro - Serviço Federal de Processamento de Dados"
				} else if strings.Contains(authStr, "ICP-Brasil") {
					timestampInfo["authority"] = "ICP-Brasil Certificate Authority"
				} else {
					if len(authStr) > 40 {
						authStr = authStr[:40] + "..."
					}
					timestampInfo["authority"] = authStr
				}
				break
			}
		}
		
		// Set default values if not found
		if timestampInfo["time"] == "" {
			timestampInfo["time"] = "Unknown"
		}
		if timestampInfo["authority"] == "" {
			timestampInfo["authority"] = "Unknown"
		}
	}

	return hasTimestamp, timestampInfo
}
func (pa *PDFAnalyzer) detectSignatureFields(ctx *model.Context, info *PDFInfo) bool {
	if ctx == nil || ctx.RootDict == nil {
		return false
	}
	
	signatureFound := false
	
	// Method 1: Look for AcroForm dictionary
	acroFormObj, found, _ := ctx.RootDict.Entry("AcroForm", "", false)
	if found && acroFormObj != nil {
		if count := pa.processAcroForm(ctx, acroFormObj, info); count > 0 {
			signatureFound = true
		}
	}
	
	// Method 2: Search for signature objects in the PDF structure
	// Look through all objects for signature-related entries
	signatureCount := 0
	
	// Check if there are any objects with /Type/Sig
	if ctx.XRefTable != nil {
		for i := 1; i <= *ctx.XRefTable.Size; i++ {
			entry, _ := ctx.XRefTable.FindTableEntry(i, 0)
			if entry == nil || entry.Object == nil {
				continue
			}
			
			if dict, ok := entry.Object.(types.Dict); ok {
				// Check for signature type
				if typeObj, found, _ := dict.Entry("Type", "", false); found {
					if nameObj, ok := typeObj.(types.Name); ok && string(nameObj) == "Sig" {
						signatureCount++
					}
				}
				
				// Check for signature field type
				if ftObj, found, _ := dict.Entry("FT", "", false); found {
					if nameObj, ok := ftObj.(types.Name); ok && string(nameObj) == "Sig" {
						signatureCount++
					}
				}
				
				// Check for signature value (V) entry that points to signature dict
				if vObj, found, _ := dict.Entry("V", "", false); found && vObj != nil {
					if indRef, ok := vObj.(types.IndirectRef); ok {
						// Try to dereference the signature value
						if sigDict, err := ctx.Dereference(indRef); err == nil {
							if sigDictTyped, ok := sigDict.(types.Dict); ok {
								// Check if this looks like a signature dictionary
								if filterObj, found, _ := sigDictTyped.Entry("Filter", "", false); found {
									if nameObj, ok := filterObj.(types.Name); ok {
										filter := string(nameObj)
										if filter == "Adobe.PPKLite" || filter == "Adobe.PPKMS" || strings.Contains(filter, "PKCS") {
											signatureCount++
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Method 3: Simple brute force search for signature patterns
	// This is a fallback when the PDF structure is not easily accessible
	if !signatureFound && signatureCount == 0 {
		// Look for signature indicators in the document catalog
		if pa.hasSignatureIndicators(ctx) {
			signatureCount = 1 // Assume at least one signature
		}
	}
	
	if signatureCount > 0 || signatureFound {
		info.HasDigitalSignatures = true
		if signatureCount > info.SignatureCount {
			info.SignatureCount = signatureCount
		}
		return true
	}
	
	return false
}

// hasSignatureIndicators checks for general signature indicators
func (pa *PDFAnalyzer) hasSignatureIndicators(ctx *model.Context) bool {
	// Check for SigFlags in the document catalog or AcroForm
	if ctx.RootDict != nil {
		// Look for any reference to signature-related entries
		if acroFormObj, found, _ := ctx.RootDict.Entry("AcroForm", "", false); found && acroFormObj != nil {
			// Even if we can't process the AcroForm fully, check for SigFlags
			if indRef, ok := acroFormObj.(types.IndirectRef); ok {
				if resolved, err := ctx.Dereference(indRef); err == nil {
					if dict, ok := resolved.(types.Dict); ok {
						if sigFlagsObj, found, _ := dict.Entry("SigFlags", "", false); found && sigFlagsObj != nil {
							return true
						}
					}
				}
			} else if dict, ok := acroFormObj.(types.Dict); ok {
				if sigFlagsObj, found, _ := dict.Entry("SigFlags", "", false); found && sigFlagsObj != nil {
					return true
				}
			}
		}
	}
	return false
}

// processAcroForm processes the AcroForm dictionary to find signature fields
func (pa *PDFAnalyzer) processAcroForm(ctx *model.Context, acroFormObj types.Object, info *PDFInfo) int {
	acroFormDict, ok := acroFormObj.(types.Dict)
	if !ok {
		return 0
	}
	
	// Look for Fields array in AcroForm
	fieldsObj, found, _ := acroFormDict.Entry("Fields", "", false)
	if !found || fieldsObj == nil {
		return 0
	}
	
	// Handle indirect references
	if indRef, ok := fieldsObj.(types.IndirectRef); ok {
		resolvedObj, err := ctx.Dereference(indRef)
		if err != nil {
			return 0
		}
		fieldsObj = resolvedObj
	}
	
	fieldsArray, ok := fieldsObj.(types.Array)
	if !ok {
		return 0
	}
	
	signatureCount := 0
	
	// Iterate through form fields
	for _, fieldRef := range fieldsArray {
		fieldDict := pa.resolveFieldDict(ctx, fieldRef)
		if fieldDict == nil {
			continue
		}
		
		// Check if this is a signature field
		if pa.isSignatureField(fieldDict) {
			signatureCount++
		}
	}
	
	if signatureCount > 0 {
		info.HasDigitalSignatures = true
		info.SignatureCount = signatureCount
	}
	
	return signatureCount
}

// detectSignaturesByteAnalysis performs raw byte analysis for signature detection
func (pa *PDFAnalyzer) detectSignaturesByteAnalysis(filePath string) (bool, int, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, 0, err
	}
	
	content := string(data)
	signatureCount := 0
	
	// Look for signature-related patterns in the raw PDF content
	patterns := []string{
		"/Type/Sig",
		"/FT/Sig", 
		"/SigFlags",
		"Adobe.PPKLite",
		"Adobe.PPKMS",
		"PKCS#7",
		"pkcs7",
		"/ByteRange",
		"/Contents<",
		"/SubFilter/adbe.pkcs7.detached",
		"/SubFilter/adbe.pkcs7.sha1",
		"/SubFilter/ETSI.CAdES.detached",
		"/Filter/Adobe.PPKLite",
		"/Filter/Adobe.PPKMS",
	}
	
	for _, pattern := range patterns {
		count := strings.Count(content, pattern)
		if count > 0 {
			if pattern == "/Type/Sig" || pattern == "/FT/Sig" {
				signatureCount += count
			} else if signatureCount == 0 {
				signatureCount = 1 // At least one signature indicated
			}
		}
	}
	
	// Additional heuristics for encrypted PDFs
	if signatureCount == 0 {
		// Look for signature dictionaries even in encrypted content
		sigDictPatterns := []string{
			"/Sig",
			"<</Type/Sig",
			"<</FT/Sig",
		}
		
		for _, pattern := range sigDictPatterns {
			if strings.Contains(content, pattern) {
				signatureCount = 1
				break
			}
		}
	}
	
	return signatureCount > 0, signatureCount, nil
}

// resolveFieldDict resolves a field reference to its dictionary
func (pa *PDFAnalyzer) resolveFieldDict(ctx *model.Context, fieldRef types.Object) types.Dict {
	// Handle indirect references
	if indRef, ok := fieldRef.(types.IndirectRef); ok {
		resolvedObj, err := ctx.Dereference(indRef)
		if err != nil {
			return nil
		}
		fieldRef = resolvedObj
	}
	
	fieldDict, ok := fieldRef.(types.Dict)
	if !ok {
		return nil
	}
	
	return fieldDict
}

// isSignatureField checks if a field dictionary represents a signature field
func (pa *PDFAnalyzer) isSignatureField(fieldDict types.Dict) bool {
	// Check field type (FT)
	ftObj, found, _ := fieldDict.Entry("FT", "", false)
	if found && ftObj != nil {
		if nameObj, ok := ftObj.(types.Name); ok {
			fieldType := string(nameObj)
			if fieldType == "Sig" {
				return true
			}
		}
	}
	
	// Check subtype or other signature indicators
	subtypeObj, found, _ := fieldDict.Entry("Subtype", "", false)
	if found && subtypeObj != nil {
		if nameObj, ok := subtypeObj.(types.Name); ok {
			subtype := string(nameObj)
			if subtype == "Widget" || subtype == "Signature" {
				// Additional check for signature-specific entries
				vObj, vFound, _ := fieldDict.Entry("V", "", false)
				sigFlagsObj, sigFound, _ := fieldDict.Entry("SigFlags", "", false)
				if (vFound && vObj != nil) || (sigFound && sigFlagsObj != nil) {
					return true
				}
			}
		}
	}
	
	// Check for signature value (V) - presence indicates signed field
	vObj, found, _ := fieldDict.Entry("V", "", false)
	if found && vObj != nil {
		// If V exists and is a dictionary, it's likely a signature
		if _, ok := vObj.(types.Dict); ok {
			return true
		}
		if indRef, ok := vObj.(types.IndirectRef); ok {
			// Indirect reference to signature dictionary
			_ = indRef // We found a signature reference
			return true
		}
	}
	
	return false
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func (pa *PDFAnalyzer) PrintReport(info *PDFInfo) {
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("                        PDF ANALYSIS REPORT")
	fmt.Println("=" + strings.Repeat("=", 80))

	// File information
	fmt.Println("\n📁 FILE INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("File name: %s\n", info.FileName)
	fmt.Printf("Path: %s\n", info.FilePath)
	fmt.Printf("Size: %s (%d bytes)\n", info.FileSizeHuman, info.FileSize)
	fmt.Printf("Last modified: %s\n", info.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("MD5: %s\n", info.MD5Hash)
	fmt.Printf("SHA256: %s\n", info.SHA256Hash)

	// Document information
	fmt.Println("\n📄 DOCUMENT METADATA")
	fmt.Println(strings.Repeat("-", 50))
	printIfNotEmpty("Title", info.Title)
	printIfNotEmpty("Author", info.Author)
	printIfNotEmpty("Subject", info.Subject)
	printIfNotEmpty("Keywords", info.Keywords)
	printIfNotEmpty("Creator", info.Creator)
	printIfNotEmpty("Producer", info.Producer)
	printIfNotEmpty("Creation date", info.CreationDate)
	printIfNotEmpty("Modification date", info.ModDate)

	// Technical information
	fmt.Println("\n⚙️  TECHNICAL INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("PDF version: %s\n", info.PDFVersion)
	fmt.Printf("Number of pages: %d\n", info.PageCount)
	fmt.Printf("Is encrypted: %s\n", boolToYesNo(info.IsEncrypted))
	fmt.Printf("Is linearized: %s\n", boolToYesNo(info.IsLinearized))
	fmt.Printf("Is tagged (accessible): %s\n", boolToYesNo(info.IsTagged))
	fmt.Printf("Has bookmarks: %s\n", boolToYesNo(info.HasBookmarks))
	fmt.Printf("Has attachments: %s\n", boolToYesNo(info.HasAttachments))
	fmt.Printf("Has forms: %s\n", boolToYesNo(info.HasForms))
	fmt.Printf("Has JavaScript: %s\n", boolToYesNo(info.HasJavaScript))
	fmt.Printf("Has annotations: %s\n", boolToYesNo(info.HasAnnotations))
	fmt.Printf("Has digital signatures: %s\n", boolToYesNo(info.HasDigitalSignatures))
	if info.HasDigitalSignatures {
		fmt.Printf("Number of signatures: %d\n", info.SignatureCount)
	}

	// Security information
	if info.IsEncrypted {
		fmt.Println("\n🔒 SECURITY INFORMATION")
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("User password set: %s\n", boolToYesNo(info.UserPasswordSet))
		fmt.Printf("Owner password set: %s\n", boolToYesNo(info.OwnerPasswordSet))
		fmt.Printf("Printing allowed: %s\n", boolToYesNo(info.PrintAllowed))
		fmt.Printf("Modification allowed: %s\n", boolToYesNo(info.ModifyAllowed))
		fmt.Printf("Copy allowed: %s\n", boolToYesNo(info.CopyAllowed))
		fmt.Printf("Add notes allowed: %s\n", boolToYesNo(info.AddNotesAllowed))
		fmt.Printf("Fill forms allowed: %s\n", boolToYesNo(info.FillFormsAllowed))
		fmt.Printf("Accessibility access: %s\n", boolToYesNo(info.AccessibilityAllowed))
		fmt.Printf("Document assembly allowed: %s\n", boolToYesNo(info.AssembleAllowed))
		fmt.Printf("High quality printing: %s\n", boolToYesNo(info.PrintHighQualityAllowed))
	}

	// Content information
	fmt.Println("\n📝 CONTENT INFORMATION")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total text characters: %d\n", info.TotalTextLength)
	fmt.Printf("Number of images: %d\n", info.ImagesCount)
	
	if len(info.FontsUsed) > 0 {
		fmt.Printf("Fonts used: %s\n", strings.Join(info.FontsUsed, ", "))
	}

	// Page information
	if len(info.Pages) > 0 {
		fmt.Println("\n📖 PAGE INFORMATION")
		fmt.Println(strings.Repeat("-", 50))
		for i, page := range info.Pages {
			if i < 5 { // Show only the first 5 pages
				fmt.Printf("Page %d: %.1f x %.1f pts, rotation: %d°, text: %d chars\n",
					page.Number, page.Width, page.Height, page.Rotation, page.TextLength)
			}
		}
		if len(info.Pages) > 5 {
			fmt.Printf("... and %d more pages\n", len(info.Pages)-5)
		}
	}

	// Bookmarks
	if len(info.Bookmarks) > 0 {
		fmt.Println("\n🔖 BOOKMARKS")
		fmt.Println(strings.Repeat("-", 50))
		for _, bookmark := range info.Bookmarks {
			indent := strings.Repeat("  ", bookmark.Level-1)
			fmt.Printf("%s- %s (page %d)\n", indent, bookmark.Title, bookmark.Page)
		}
	}

	// Attachments
	if len(info.Attachments) > 0 {
		fmt.Println("\n📎 ATTACHMENTS")
		fmt.Println(strings.Repeat("-", 50))
		for _, attachment := range info.Attachments {
			fmt.Printf("- %s (%s, %s)\n", attachment.Name, attachment.Type, formatFileSize(attachment.Size))
		}
	}

	// Digital signatures - always visible section
	fmt.Println("\n🔐 DIGITAL SIGNATURES")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Document has signatures: %s\n", boolToYesNo(info.HasDigitalSignatures))
	fmt.Printf("Number of signatures: %d\n", info.SignatureCount)
	
	if info.HasDigitalSignatures && len(info.Signatures) > 0 {
		fmt.Println("\nSignature details:")
		for i, sig := range info.Signatures {
			fmt.Printf("\n  Signature %d:\n", i+1)
			if sig.FieldName != "" {
				fmt.Printf("    Field: %s\n", sig.FieldName)
			}
			fmt.Printf("    Type: %s\n", sig.Type)
			if sig.SubFilter != "" {
				fmt.Printf("    SubFilter: %s\n", sig.SubFilter)
			}
			fmt.Printf("    Status: %s\n", sig.Status)
			fmt.Printf("    Valid: %s\n", boolToYesNo(sig.IsValid))
			fmt.Printf("    Certified: %s\n", boolToYesNo(sig.IsCertified))
			if sig.SignerName != "" {
				fmt.Printf("    Signer: %s\n", sig.SignerName)
			}
			if sig.SigningTime != "" {
				fmt.Printf("    Signing date/time: %s\n", sig.SigningTime)
			}
			if sig.Location != "" {
				fmt.Printf("    Location: %s\n", sig.Location)
			}
			if sig.Reason != "" {
				fmt.Printf("    Reason: %s\n", sig.Reason)
			}
			if sig.ContactInfo != "" {
				fmt.Printf("    Contact: %s\n", sig.ContactInfo)
			}
			
			// Timestamp information
			fmt.Printf("    Has timestamp: %s\n", boolToYesNo(sig.HasTimestamp))
			if sig.HasTimestamp {
				if sig.TimestampType != "" {
					fmt.Printf("    Timestamp type: %s\n", sig.TimestampType)
				}
				if sig.TimestampTime != "" {
					fmt.Printf("    Timestamp time: %s\n", sig.TimestampTime)
				}
				if sig.TimestampAuthority != "" {
					fmt.Printf("    Timestamp authority: %s\n", sig.TimestampAuthority)
				}
				if sig.TimestampStatus != "" {
					fmt.Printf("    Timestamp status: %s\n", sig.TimestampStatus)
				}
			}
			
			if len(sig.ValidationErrors) > 0 {
				fmt.Printf("    Validation issues:\n")
				for _, err := range sig.ValidationErrors {
					fmt.Printf("      - %s\n", err)
				}
			}
		}
	} else if info.HasDigitalSignatures {
		fmt.Println("\nDigital signature(s) detected in document.")
		fmt.Printf("Found %d signature(s), but detailed validation failed due to encryption or security restrictions.\n", info.SignatureCount)
		fmt.Println("\nSignature validation requires:")
		fmt.Println("  • Document decryption (if encrypted)")
		fmt.Println("  • Access to signing certificates")
		fmt.Println("  • Valid certificate chain")
		fmt.Println("  • Trusted certificate authority (CA)")
	} else {
		fmt.Println("\nThis document does not have digital signatures.")
		fmt.Println("To digitally sign a PDF, you can use:")
		fmt.Println("  • Adobe Acrobat")
		fmt.Println("  • LibreOffice")
		fmt.Println("  • Online signature tools")
		fmt.Println("  • ICP-Brasil digital certificates")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Analysis completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 80))
}

// Funções utilitárias
func getStringFromDict(dict types.Dict, key string) string {
	if obj, found := dict.Find(key); found && obj != nil {
		if str, ok := obj.(types.StringLiteral); ok {
			return str.Value()
		}
		if name, ok := obj.(types.Name); ok {
			return name.Value()
		}
		// Também tentar como HexLiteral
		if hex, ok := obj.(types.HexLiteral); ok {
			return hex.Value()
		}
	}
	return ""
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func printIfNotEmpty(label, value string) {
	if value != "" {
		fmt.Printf("%s: %s\n", label, value)
	}
}