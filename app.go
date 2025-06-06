package main

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
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
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run . <caminho_do_pdf>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]
	
	analyzer := &PDFAnalyzer{}
	info, err := analyzer.AnalyzePDF(pdfPath)
	if err != nil {
		log.Fatalf("Erro ao analisar PDF: %v", err)
	}

	analyzer.PrintReport(info)
}

type PDFAnalyzer struct{}

func (pa *PDFAnalyzer) AnalyzePDF(filePath string) (*PDFInfo, error) {
	info := &PDFInfo{}
	
	// Informações básicas do arquivo
	if err := pa.getFileInfo(filePath, info); err != nil {
		return nil, fmt.Errorf("erro ao obter informações do arquivo: %v", err)
	}

	// Análise usando pdfcpu
	if err := pa.analyzePDFCPU(filePath, info); err != nil {
		fmt.Printf("Aviso: erro na análise pdfcpu: %v\n", err)
	}

	// Análise usando ledongthuc/pdf
	if err := pa.analyzeLedongthuc(filePath, info); err != nil {
		fmt.Printf("Aviso: erro na análise ledongthuc: %v\n", err)
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
			fmt.Printf("Aviso: não foi possível desreferenciar o dicionário Info: %v\\n", err)
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
				fmt.Printf("Aviso: Info object não é um dicionário, mas sim do tipo %T\\n", infoObject)
			}
		}
	}

	// Informações técnicas
	if ctx.HeaderVersion != nil {
		info.PDFVersion = fmt.Sprintf("%.1f", float64(*ctx.HeaderVersion))
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
	// Validar assinaturas usando pdfcpu
	results, err := api.ValidateSignaturesFile(filePath, true, true, nil) // all=true, full=true
	if err != nil {
		fmt.Printf("Aviso: erro ao validar assinaturas: %v\n", err)
		info.HasDigitalSignatures = false
		info.SignatureCount = 0
		return
	}

	if len(results) == 0 {
		info.HasDigitalSignatures = false
		info.SignatureCount = 0
		return
	}

	info.HasDigitalSignatures = true
	info.SignatureCount = len(results)
	info.Signatures = make([]DigitalSignatureInfo, 0, len(results))

	// Processar cada resultado de validação
	for i, result := range results {
		sigInfo := DigitalSignatureInfo{
			FieldName: fmt.Sprintf("Signature_%d", i+1),
			Status:    "Processada",
			IsValid:   true, // Assumir válida se não houver erro na validação
			Type:      "Digital",
		}

		// Analisar o resultado da validação (string)
		if strings.Contains(strings.ToLower(result), "invalid") {
			sigInfo.Status = "Inválida"
			sigInfo.IsValid = false
		} else if strings.Contains(strings.ToLower(result), "valid") {
			sigInfo.Status = "Válida"
			sigInfo.IsValid = true
		} else {
			sigInfo.Status = "Desconhecida"
			sigInfo.IsValid = false
		}

		// Extrair informações básicas da string de resultado
		if strings.Contains(result, "certified") {
			sigInfo.Type = "Certificada"
			sigInfo.IsCertified = true
		} else {
			sigInfo.Type = "Aprovação"
			sigInfo.IsCertified = false
		}

		// Adicionar detalhes da validação
		sigInfo.ValidationErrors = []string{result}

		info.Signatures = append(info.Signatures, sigInfo)
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func (pa *PDFAnalyzer) PrintReport(info *PDFInfo) {
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("                        RELATÓRIO DE ANÁLISE DE PDF")
	fmt.Println("=" + strings.Repeat("=", 80))

	// Informações do arquivo
	fmt.Println("\n📁 INFORMAÇÕES DO ARQUIVO")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Nome do arquivo: %s\n", info.FileName)
	fmt.Printf("Caminho: %s\n", info.FilePath)
	fmt.Printf("Tamanho: %s (%d bytes)\n", info.FileSizeHuman, info.FileSize)
	fmt.Printf("Última modificação: %s\n", info.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("MD5: %s\n", info.MD5Hash)
	fmt.Printf("SHA256: %s\n", info.SHA256Hash)

	// Informações do documento
	fmt.Println("\n📄 METADADOS DO DOCUMENTO")
	fmt.Println(strings.Repeat("-", 50))
	printIfNotEmpty("Título", info.Title)
	printIfNotEmpty("Autor", info.Author)
	printIfNotEmpty("Assunto", info.Subject)
	printIfNotEmpty("Palavras-chave", info.Keywords)
	printIfNotEmpty("Criador", info.Creator)
	printIfNotEmpty("Producer", info.Producer)
	printIfNotEmpty("Data de criação", info.CreationDate)
	printIfNotEmpty("Data de modificação", info.ModDate)

	// Informações técnicas
	fmt.Println("\n⚙️  INFORMAÇÕES TÉCNICAS")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Versão PDF: %s\n", info.PDFVersion)
	fmt.Printf("Número de páginas: %d\n", info.PageCount)
	fmt.Printf("Está criptografado: %s\n", boolToYesNo(info.IsEncrypted))
	fmt.Printf("É linearizado: %s\n", boolToYesNo(info.IsLinearized))
	fmt.Printf("É tagged (acessível): %s\n", boolToYesNo(info.IsTagged))
	fmt.Printf("Tem marcadores: %s\n", boolToYesNo(info.HasBookmarks))
	fmt.Printf("Tem anexos: %s\n", boolToYesNo(info.HasAttachments))
	fmt.Printf("Tem formulários: %s\n", boolToYesNo(info.HasForms))
	fmt.Printf("Tem JavaScript: %s\n", boolToYesNo(info.HasJavaScript))
	fmt.Printf("Tem anotações: %s\n", boolToYesNo(info.HasAnnotations))
	fmt.Printf("Tem assinaturas digitais: %s\n", boolToYesNo(info.HasDigitalSignatures))
	if info.HasDigitalSignatures {
		fmt.Printf("Número de assinaturas: %d\n", info.SignatureCount)
	}

	// Informações de segurança
	if info.IsEncrypted {
		fmt.Println("\n🔒 INFORMAÇÕES DE SEGURANÇA")
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("Senha de usuário definida: %s\n", boolToYesNo(info.UserPasswordSet))
		fmt.Printf("Senha de proprietário definida: %s\n", boolToYesNo(info.OwnerPasswordSet))
		fmt.Printf("Impressão permitida: %s\n", boolToYesNo(info.PrintAllowed))
		fmt.Printf("Modificação permitida: %s\n", boolToYesNo(info.ModifyAllowed))
		fmt.Printf("Cópia permitida: %s\n", boolToYesNo(info.CopyAllowed))
		fmt.Printf("Adicionar notas permitido: %s\n", boolToYesNo(info.AddNotesAllowed))
		fmt.Printf("Preencher formulários permitido: %s\n", boolToYesNo(info.FillFormsAllowed))
		fmt.Printf("Acesso para acessibilidade: %s\n", boolToYesNo(info.AccessibilityAllowed))
		fmt.Printf("Montagem de documento permitida: %s\n", boolToYesNo(info.AssembleAllowed))
		fmt.Printf("Impressão alta qualidade: %s\n", boolToYesNo(info.PrintHighQualityAllowed))
	}

	// Informações de conteúdo
	fmt.Println("\n📝 INFORMAÇÕES DE CONTEÚDO")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total de caracteres de texto: %d\n", info.TotalTextLength)
	fmt.Printf("Número de imagens: %d\n", info.ImagesCount)
	
	if len(info.FontsUsed) > 0 {
		fmt.Printf("Fontes utilizadas: %s\n", strings.Join(info.FontsUsed, ", "))
	}

	// Informações das páginas
	if len(info.Pages) > 0 {
		fmt.Println("\n📖 INFORMAÇÕES DAS PÁGINAS")
		fmt.Println(strings.Repeat("-", 50))
		for i, page := range info.Pages {
			if i < 5 { // Mostrar apenas as primeiras 5 páginas
				fmt.Printf("Página %d: %.1f x %.1f pts, rotação: %d°, texto: %d chars\n",
					page.Number, page.Width, page.Height, page.Rotation, page.TextLength)
			}
		}
		if len(info.Pages) > 5 {
			fmt.Printf("... e mais %d páginas\n", len(info.Pages)-5)
		}
	}

	// Bookmarks
	if len(info.Bookmarks) > 0 {
		fmt.Println("\n🔖 MARCADORES")
		fmt.Println(strings.Repeat("-", 50))
		for _, bookmark := range info.Bookmarks {
			indent := strings.Repeat("  ", bookmark.Level-1)
			fmt.Printf("%s- %s (página %d)\n", indent, bookmark.Title, bookmark.Page)
		}
	}

	// Anexos
	if len(info.Attachments) > 0 {
		fmt.Println("\n📎 ANEXOS")
		fmt.Println(strings.Repeat("-", 50))
		for _, attachment := range info.Attachments {
			fmt.Printf("- %s (%s, %s)\n", attachment.Name, attachment.Type, formatFileSize(attachment.Size))
		}
	}

	// Assinaturas digitais
	if info.HasDigitalSignatures && len(info.Signatures) > 0 {
		fmt.Println("\n🔐 ASSINATURAS DIGITAIS")
		fmt.Println(strings.Repeat("-", 50))
		for i, sig := range info.Signatures {
			fmt.Printf("Assinatura %d:\n", i+1)
			if sig.FieldName != "" {
				fmt.Printf("  Campo: %s\n", sig.FieldName)
			}
			fmt.Printf("  Tipo: %s\n", sig.Type)
			fmt.Printf("  SubFilter: %s\n", sig.SubFilter)
			fmt.Printf("  Status: %s\n", sig.Status)
			fmt.Printf("  Válida: %s\n", boolToYesNo(sig.IsValid))
			fmt.Printf("  Certificada: %s\n", boolToYesNo(sig.IsCertified))
			if sig.SignerName != "" {
				fmt.Printf("  Assinante: %s\n", sig.SignerName)
			}
			if sig.SigningTime != "" {
				fmt.Printf("  Data/Hora da assinatura: %s\n", sig.SigningTime)
			}
			if sig.Location != "" {
				fmt.Printf("  Local: %s\n", sig.Location)
			}
			if sig.Reason != "" {
				fmt.Printf("  Motivo: %s\n", sig.Reason)
			}
			if sig.ContactInfo != "" {
				fmt.Printf("  Contato: %s\n", sig.ContactInfo)
			}
			if len(sig.ValidationErrors) > 0 {
				fmt.Printf("  Problemas de validação:\n")
				for _, err := range sig.ValidationErrors {
					fmt.Printf("    - %s\n", err)
				}
			}
			if i < len(info.Signatures)-1 {
				fmt.Println()
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Análise concluída em: %s\n", time.Now().Format("2006-01-02 15:04:05"))
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
		return "Sim"
	}
	return "Não"
}

func printIfNotEmpty(label, value string) {
	if value != "" {
		fmt.Printf("%s: %s\n", label, value)
	}
}