package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Uso: go run debug_version.go <arquivo.pdf>")
	}

	filename := os.Args[1]
	
	// Ler o contexto do PDF
	ctx, err := api.ReadContextFile(filename)
	if err != nil {
		log.Fatalf("Erro ao ler o PDF: %v", err)
	}

	fmt.Printf("Raw HeaderVersion: %v\n", ctx.HeaderVersion)
	if ctx.HeaderVersion != nil {
		fmt.Printf("HeaderVersion value: %d\n", *ctx.HeaderVersion)
		fmt.Printf("Divided by 10: %.1f\n", float64(*ctx.HeaderVersion)/10.0)
		fmt.Printf("Alternative formats:\n")
		fmt.Printf("  As string: %s\n", fmt.Sprintf("%d", *ctx.HeaderVersion))
		fmt.Printf("  First digit + . + second digit: %d.%d\n", *ctx.HeaderVersion/10, *ctx.HeaderVersion%10)
		
		// Tenta diferentes interpretações
		val := *ctx.HeaderVersion
		if val >= 10 {
			major := val / 10
			minor := val % 10
			fmt.Printf("  Interpretation: %d.%d\n", major, minor)
		}
	}
}
