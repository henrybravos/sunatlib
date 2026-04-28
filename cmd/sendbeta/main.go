// cmd/sendbeta/main.go — Envía los XMLs de testdata/ al entorno beta de SUNAT.
// Firma digitalmente si se provee certificado, o envía sin firma (beta lo acepta igual).
//
// Uso:
//   SUNAT_RUC=20XXXXXXXXX SUNAT_CERT_PASS=mipassword go run ./cmd/sendbeta
//
// Variables de entorno:
//   SUNAT_RUC        RUC del emisor (requerido)
//   SUNAT_CERT_PASS  Password del archivo .p12 (requerido para firma)
//   SUNAT_CERT_PATH  Ruta al .p12 (default: /Users/hbs/Documents/infira/infac/certificados/infira.p12)
//   SUNAT_XML        Ruta a un XML específico (default: todos en testdata/)
package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/henrybravos/sunatlib"
)

const defaultCertPath = "/Users/hbs/Documents/infira/infac/certificados/infira.p12"

type minDoc struct {
	ID                 string `xml:"ID"`
	InvoiceTypeCode    string `xml:"InvoiceTypeCode"`
	CreditNoteTypeCode string `xml:"CreditNoteTypeCode"`
	DebitNoteTypeCode  string `xml:"DebitNoteTypeCode"`
}

func main() {
	// Cargar .env si existe (gitignored, nunca se commitea)
	loadDotEnv(".env")

	ruc := getenv("SUNAT_RUC", "")
	if ruc == "" {
		fmt.Println("❌ Falta SUNAT_RUC — Ej: SUNAT_RUC=20XXXXXXXXX go run ./cmd/sendbeta")
		os.Exit(1)
	}

	certPath := getenv("SUNAT_CERT_PATH", defaultCertPath)
	certPass := getenv("SUNAT_CERT_PASS", "")

	endpoint := sunatlib.GetBillServiceEndpoint(sunatlib.Beta)
	client := sunatlib.NewSUNATClient(ruc, "MODDATOS", "MODDATOS", endpoint)

	// Configurar certificado si el password está disponible
	signed := false
	if certPass != "" {
		if _, err := os.Stat(certPath); err == nil {
			if err := client.SetCertificateFromPFX(certPath, certPass, os.TempDir()); err != nil {
				fmt.Printf("⚠️  No se pudo cargar el certificado: %v\n", err)
				fmt.Println("   Continuando sin firma (beta acepta sin firma)...")
			} else {
				signed = true
				fmt.Printf("🔑 Certificado cargado: %s\n", certPath)
			}
		} else {
			fmt.Printf("⚠️  Certificado no encontrado en: %s\n", certPath)
		}
	} else {
		fmt.Println("ℹ️  SUNAT_CERT_PASS no definido — enviando sin firma digital")
	}

	// Determinar qué archivos enviar
	xmlPath := getenv("SUNAT_XML", "")
	var files []string
	if xmlPath != "" {
		files = []string{xmlPath}
	} else {
		var err error
		files, err = filepath.Glob(filepath.Join("testdata", "F*.xml"))
		if err != nil || len(files) == 0 {
			fmt.Println("❌ No se encontraron archivos XML en testdata/")
			os.Exit(1)
		}
	}

	fmt.Printf("\n🌐 Endpoint: %s\n", endpoint)
	fmt.Printf("👤 RUC: %s | Firma: %v\n", ruc, signed)

	for _, f := range files {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📄 Procesando: %s\n", filepath.Base(f))

		xmlContent, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("  ❌ Error leyendo: %v\n", err)
			continue
		}

		// Validar estructura UBL
		if err := client.ValidateUBL(xmlContent); err != nil {
			fmt.Printf("  ⚠️  Validación UBL: %v\n", err)
		} else {
			fmt.Printf("  ✅ Validación UBL: OK\n")
		}

		docID, docType := extractDocInfo(xmlContent)
		if docID == "" {
			fmt.Printf("  ❌ No se pudo extraer ID del documento\n")
			continue
		}
		fmt.Printf("  📋 Documento: %s | Tipo: %s\n", docID, docType)

		// Sustituir el RUC placeholder en los XMLs de testdata
		xmlContent = substituteRUC(xmlContent, ruc)

		var payload []byte
		if signed {
			payload, err = client.SignXML(xmlContent)
			if err != nil {
				fmt.Printf("  ⚠️  Error firmando (se envía sin firma): %v\n", err)
				payload = xmlContent
			} else {
				fmt.Printf("  ✍️  XML firmado\n")
			}
		} else {
			payload = xmlContent
		}

		resp, err := client.SendToSUNAT(payload, docType, docID)
		if err != nil {
			fmt.Printf("  ❌ Error HTTP: %v\n", err)
			continue
		}

		if resp.Success {
			fmt.Printf("  ✅ SUNAT ACEPTÓ\n")
		} else {
			fmt.Printf("  ❌ SUNAT RECHAZÓ\n")
		}
		fmt.Printf("  💬 %s\n", resp.Message)

		// Mostrar respuesta cruda (primeros 800 chars)
		if len(resp.ResponseXML) > 0 {
			raw := string(resp.ResponseXML)
			if len(raw) > 800 {
				raw = raw[:800] + "\n... (truncado)"
			}
			fmt.Printf("  📩 Respuesta:\n%s\n", indent(raw))
		}

		// Pausa entre requests para evitar rate limiting de SUNAT beta
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Println("🏁 Finalizado.")

	// Limpiar temp del signer
	client.Cleanup()
}

// substituteRUC reemplaza el RUC placeholder en los XMLs de testdata
// con el RUC real del emisor antes de firmar y enviar.
func substituteRUC(xmlContent []byte, ruc string) []byte {
	const placeholder = "20000000001"
	return []byte(strings.ReplaceAll(string(xmlContent), placeholder, ruc))
}

// loadDotEnv carga variables desde un archivo .env (KEY=VALUE).
// Solo setea si la variable no está ya definida en el entorno.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env es opcional
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// No sobreescribir variables ya seteadas en el entorno
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "     " + l
	}
	return strings.Join(lines, "\n")
}

func extractDocInfo(xmlContent []byte) (docID, docType string) {
	var doc minDoc
	if err := xml.Unmarshal(xmlContent, &doc); err != nil {
		return extractByString(xmlContent)
	}
	docID = doc.ID
	switch {
	case doc.InvoiceTypeCode != "":
		docType = doc.InvoiceTypeCode
	case doc.CreditNoteTypeCode != "":
		docType = "07"
	case doc.DebitNoteTypeCode != "":
		docType = "08"
	default:
		docType = inferDocType(docID)
	}
	return docID, docType
}

func extractByString(xmlContent []byte) (docID, docType string) {
	s := string(xmlContent)
	start := strings.Index(s, "<cbc:ID>")
	if start == -1 {
		return "", ""
	}
	start += len("<cbc:ID>")
	end := strings.Index(s[start:], "</cbc:ID>")
	if end == -1 {
		return "", ""
	}
	docID = strings.TrimSpace(s[start : start+end])

	tcStart := strings.Index(s, "<cbc:InvoiceTypeCode")
	if tcStart != -1 {
		gtPos := strings.Index(s[tcStart:], ">")
		if gtPos != -1 {
			rest := s[tcStart+gtPos+1:]
			closeEnd := strings.Index(rest, "<")
			if closeEnd != -1 {
				docType = strings.TrimSpace(rest[:closeEnd])
			}
		}
	}
	if docType == "" {
		docType = inferDocType(docID)
	}
	return docID, docType
}

func inferDocType(docID string) string {
	if len(docID) == 0 {
		return "01"
	}
	switch docID[0] {
	case 'F', 'f':
		return "01"
	case 'B', 'b':
		return "03"
	default:
		return "01"
	}
}
