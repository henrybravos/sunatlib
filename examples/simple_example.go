package main

import (
	"fmt"
	"log"
	"os"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Example usage of SUNATLib
	err := sendInvoiceExample()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func sendInvoiceExample() error {
	// Create SUNAT client with your credentials
	client := sunatlib.NewSUNATClient(
		"20123456789",     // Your RUC number
		"MODDATOS",        // SOL username
		"moddatos",        // SOL password  
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", // Beta endpoint
	)
	defer client.Cleanup()

	// Configure certificate from PFX file
	err := client.SetCertificateFromPFX(
		"./certificate.pfx",    // Path to your certificate
		"your_cert_password",   // Certificate password
		"/tmp/sunatlib_certs",  // Temporary directory
	)
	if err != nil {
		return fmt.Errorf("failed to configure certificate: %w", err)
	}

	// Read invoice XML file
	xmlContent, err := os.ReadFile("./invoice.xml")
	if err != nil {
		return fmt.Errorf("failed to read XML file: %w", err)
	}

	// Step 1: Sign the XML document
	fmt.Println("🔏 Signing XML document...")
	signedXML, err := client.SignXML(xmlContent)
	if err != nil {
		return fmt.Errorf("failed to sign XML: %w", err)
	}
	fmt.Printf("✅ XML signed successfully (%d bytes)\n", len(signedXML))

	// Optional: Save signed XML for inspection or later use
	err = os.WriteFile("invoice_signed.xml", signedXML, 0644)
	if err != nil {
		log.Printf("⚠️ Failed to save signed XML: %v", err)
	} else {
		fmt.Println("💾 Signed XML saved as: invoice_signed.xml")
	}

	// Step 2: User decides when to send to SUNAT
	fmt.Println("🚀 Sending to SUNAT...")
	response, err := client.SendToSUNAT(
		signedXML,       // Previously signed XML
		"01",           // Document type (invoice)
		"F001-00000001", // Series-number
	)
	if err != nil {
		return fmt.Errorf("failed to send to SUNAT: %w", err)
	}

	// Process response
	fmt.Printf("✅ Success: %t\n", response.Success)
	fmt.Printf("📝 Message: %s\n", response.Message)

	if response.Success && response.ApplicationResponse != nil {
		err = response.SaveApplicationResponse("cdr_response.zip")
		if err != nil {
			log.Printf("⚠️ Failed to save CDR: %v", err)
		} else {
			fmt.Println("💾 CDR saved as: cdr_response.zip")
		}
	}

	return nil
}