package main

import (
	"time"
)

// PDFInfo holds comprehensive information about a PDF file
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

// PageInfo holds information about a specific page
type PageInfo struct {
	Number     int
	Width      float64
	Height     float64
	Rotation   int
	TextLength int
	ImageCount int
}

// BookmarkInfo holds information about a bookmark
type BookmarkInfo struct {
	Title string
	Level int
	Page  int
}

// AttachmentInfo holds information about an attachment
type AttachmentInfo struct {
	Name string
	Size int64
	Type string
}

// AnnotationInfo holds information about an annotation
type AnnotationInfo struct {
	Type    string
	Page    int
	Content string
}

// DigitalSignatureInfo holds information about a digital signature
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

// PDFAnalyzer is the main analyzer struct
type PDFAnalyzer struct{}
