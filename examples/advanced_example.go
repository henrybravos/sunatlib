package main

import (
	"fmt"
	"log"
	"os"

	"github.com/henrybravos/sunatlib"
	"github.com/henrybravos/sunatlib/utils"
)

func main() {
	// Ejemplo avanzado con manejo de certificados PEM

	// 1. Verificar disponibilidad de xmlsec1
	if err := utils.CheckXMLSec1Available(); err != nil {
		log.Fatalf("xmlsec1 no disponible: %v", err)
	}
	fmt.Println("✓ xmlsec1 disponible")

	// 2. Extract certificate from PFX
	pfxPath := "../../certificate.pfx"
	pfxPassword := "your_cert_password"
	tempDir := "/tmp/sunatlib_advanced"

	privateKeyPath, certPath, err := utils.ExtractPEMFromPFX(pfxPath, pfxPassword, tempDir)
	if err != nil {
		log.Fatalf("Error extrayendo certificado: %v", err)
	}
	fmt.Printf("✓ Certificado extraído:\n  - Clave privada: %s\n  - Certificado: %s\n", privateKeyPath, certPath)

	// 3. Validar certificado
	cert, err := utils.ValidateCertificate(certPath)
	if err != nil {
		log.Fatalf("Error validando certificado: %v", err)
	}
	fmt.Printf("✓ Certificado válido para: %s\n", cert.Subject.CommonName)

	// 4. Obtener información del certificado
	certInfo, err := utils.GetCertificateInfo(certPath)
	if err != nil {
		log.Fatalf("Error obteniendo info del certificado: %v", err)
	}
	fmt.Println("📋 Información del certificado:")
	for key, value := range certInfo {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// 5. Configure SUNAT client
	client := sunatlib.NewSUNATClient("20123456789", "MODDATOS", "moddatos",
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService")
	defer client.Cleanup()

	// 6. Configurar certificado con archivos PEM extraídos
	err = client.SetCertificate(privateKeyPath, certPath)
	if err != nil {
		log.Fatalf("Error configurando certificado: %v", err)
	}
	fmt.Println("✓ Cliente configurado")

	// 7. Read and process XML
	xmlContent, err := os.ReadFile("../../invoice.xml")
	if err != nil {
		log.Fatalf("Error reading XML: %v", err)
	}
	fmt.Printf("✓ XML read (%d bytes)\n", len(xmlContent))

	// 8. Sign and send
	fmt.Println("🚀 Sending to SUNAT...")
	response, err := client.SignAndSendInvoice(xmlContent, "01", "F001-00000001")
	if err != nil {
		log.Fatalf("Error enviando factura: %v", err)
	}

	// 9. Procesar respuesta
	fmt.Printf("\n📋 Resultado:\n")
	fmt.Printf("  Éxito: %t\n", response.Success)
	fmt.Printf("  Mensaje: %s\n", response.Message)

	if response.Success {
		fmt.Println("✅ Factura aceptada por SUNAT")
		
		if response.ApplicationResponse != nil {
			cdrFile := "cdr_advanced.zip"
			err = response.SaveApplicationResponse(cdrFile)
			if err != nil {
				log.Printf("Error guardando CDR: %v", err)
			} else {
				fmt.Printf("📄 CDR guardado: %s\n", cdrFile)
			}
		}
	} else {
		fmt.Println("❌ Factura rechazada")
	}

	// 10. Limpiar archivos temporales
	fmt.Println("🧹 Limpiando archivos temporales...")
}