package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/henrybravos/sunatlib"
	"github.com/henrybravos/sunatlib/gre"
	"github.com/henrybravos/sunatlib/utils"
)

func main() {
	// Load environment
	ruc := os.Getenv("SUNAT_RUC")
	certPath := os.Getenv("SUNAT_CERT_PATH")
	certPass := os.Getenv("SUNAT_CERT_PASS")
	
	clientID := os.Getenv("SUNAT_GRE_CLIENT_ID")
	clientSecret := os.Getenv("SUNAT_GRE_CLIENT_SECRET")
	solUser := os.Getenv("SUNAT_SOL_USUARIO")
	solPass := os.Getenv("SUNAT_SOL_CLAVE")

	if ruc == "" {
		fmt.Println("❌ SUNAT_RUC is required")
		os.Exit(1)
	}

	if clientID == "" || clientSecret == "" {
		fmt.Println("⚠️  SUNAT_GRE_CLIENT_ID or SUNAT_GRE_CLIENT_SECRET missing.")
		fmt.Printf("   Trying with default Beta credentials for RUC %s...\n", ruc)
		clientID = ruc + "MODDATOS"
		clientSecret = "moddatos"
	}

	// 1. Initialize SUNAT client for signing
	sunatClient := sunatlib.NewSUNATClient(ruc, "MODDATOS", "MODDATOS", sunatlib.GetBillServiceEndpoint(sunatlib.Beta))
	if certPass != "" && certPath != "" {
		// Try to see if there are PEM files first (more reliable in Go)
		keyPath := "/Users/hbs/Documents/infira/sunatlib/certs/infira.key"
		crtPath := "/Users/hbs/Documents/infira/sunatlib/certs/infira.crt"
		
		if _, err := os.Stat(keyPath); err == nil {
			fmt.Println("🔑 Loading certificate from PEM files...")
			if err := sunatClient.SetCertificate(keyPath, crtPath); err != nil {
				fmt.Printf("⚠️  Failed to load PEM: %v\n", err)
			} else {
				fmt.Println("✅ Certificate loaded from PEM.")
			}
		} else if err := sunatClient.SetCertificateFromPFX(certPath, certPass, os.TempDir()); err != nil {
			fmt.Printf("⚠️  Failed to load certificate: %v\n", err)
		} else {
			fmt.Println("🔑 Certificate loaded from PFX.")
		}
	} else {
		fmt.Println("⚠️  No certificate provided. Submission might fail if SUNAT requires signed XML.")
	}

	// 2. Initialize GRE client pointing to NubeFact Sandbox for stability
	greClient := &gre.GreClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     ruc + solUser,
		Password:     solPass,
		TokenURL:     sunatlib.NubeFactGREToken,
		ApiURL:       sunatlib.NubeFactGREApi,
	}

	// For NubeFact we don't need the magic resolver, but we keep a standard timeout
	greClient.HttpClient = &http.Client{
		Timeout: 60 * time.Second,
	}

	// 3. Get OAuth Token
	fmt.Printf("🌐 Getting OAuth token from NubeFact for %s...\n", clientID)
	_, err := greClient.GetToken(context.Background())
	if err != nil {
		fmt.Printf("❌ OAuth error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Token obtained.")

	// 4. Process examples
	examples, _ := filepath.Glob(filepath.Join("testdata", "gre", "examples", "*.xml"))
	if len(examples) == 0 {
		fmt.Println("❌ No GRE examples found in testdata/gre/examples/")
		os.Exit(1)
	}

	for i, f := range examples {
		fmt.Printf("\n--- Processing %s ---\n", filepath.Base(f))
		xmlContent, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("❌ Read error: %v\n", err)
			continue
		}

		// Substitute RUC
		xmlContent = []byte(strings.ReplaceAll(string(xmlContent), "20600000000", ruc))

		// Sign XML
		signedXML, err := sunatClient.SignXML(xmlContent)
		if err != nil {
			fmt.Printf("⚠️  Sign error (sending unsigned): %v\n", err)
			signedXML = xmlContent
		} else {
			fmt.Println("✍️  XML signed.")
		}

		// 3. Prepare transmission data
		// SUNAT REST API expects the filename in the URL to be: {RUC}-{TIPO}-{SERIE}-{NUMERO}
		// without the .zip extension, though the payload contains the ZIP.
		base := filepath.Base(f)
		_ = strings.TrimSuffix(base, ".xml")
		
		sunatFileName := fmt.Sprintf("%s-09-T001-%08d", ruc, i+1)
		
		zipFileName := sunatFileName + ".zip"
		xmlFileName := sunatFileName + ".xml"
		
		zipContent, err := utils.CreateZip(xmlFileName, signedXML)
		if err != nil {
			fmt.Printf("❌ Zip error: %v\n", err)
			continue
		}
 
		// Send it
		fmt.Printf("🚀 Sending %s to Beta...\n", zipFileName)
		resp, err := greClient.SendGuide(context.Background(), sunatFileName, zipContent)
		if err != nil {
			fmt.Printf("❌ Send error: %v\n", err)
			continue
		}

		fmt.Printf("✅ Success! Ticket: %s\n", resp.NumTicket)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n🏁 Done.")
}
