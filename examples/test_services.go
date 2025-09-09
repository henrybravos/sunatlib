package main

import (
	"fmt"
	"log"
	"os"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Create independent consultation client
	fmt.Println("=== Probando servicios de consulta (Independientes) ===")
	
	// Full consultation client (RUC + DNI)
	client := sunatlib.NewConsultationClient(os.Getenv("DECOLECTA_API_KEY"))
	
	// Or use specific clients:
	// rucClient := sunatlib.NewRUCConsultationClient(os.Getenv("DECOLECTA_API_KEY"))
	// dniClient := sunatlib.NewDNIConsultationClient() // Free
	
	rucResult, err := client.ConsultRUC("20601030013")
	if err != nil {
		log.Printf("Error consultando RUC: %v", err)
	} else if rucResult.Success {
		fmt.Printf("✅ RUC Básico: %s (%s)\n", rucResult.Data.RazonSocial, rucResult.Data.Estado)
		fmt.Printf("   Dirección: %s, %s, %s\n", rucResult.Data.Distrito, rucResult.Data.Provincia, rucResult.Data.Departamento)
	} else {
		fmt.Printf("❌ Error RUC: %s\n", rucResult.Message)
	}

	// Test RUC Full
	fmt.Println("\n=== Probando RUC Full ===")
	rucFullResult, err := client.ConsultRUCFull("20601030013")
	if err != nil {
		log.Printf("Error consultando RUC Full: %v", err)
	} else if rucFullResult.Success {
		fmt.Printf("✅ RUC Completo: %s\n", rucFullResult.Data.RazonSocial)
	} else {
		fmt.Printf("❌ Error RUC Full: %s\n", rucFullResult.Message)
	}

	// Test DNI service (free)
	fmt.Println("\n=== Probando servicio DNI (EsSalud - Gratuito) ===")
	
	dniResult, err := client.ConsultDNI("70408005")
	if err != nil {
		log.Printf("Error consultando DNI: %v", err)
	} else if dniResult.Success {
		fmt.Printf("✅ DNI: %s\n", dniResult.Data.DNI)
		fmt.Printf("   Nombre Completo: %s\n", dniResult.Data.NombreCompleto)
		fmt.Printf("   Nombres: %s\n", dniResult.Data.Nombres)
		fmt.Printf("   Apellido Paterno: %s\n", dniResult.Data.ApellidoPaterno)
		fmt.Printf("   Apellido Materno: %s\n", dniResult.Data.ApellidoMaterno)
	} else {
		fmt.Printf("❌ Error DNI: %s\n", dniResult.Message)
	}

	// Test validation functions
	fmt.Println("\n=== Probando funciones de validación ===")
	fmt.Printf("RUC válido: %v\n", sunatlib.IsValidRUC("20601030013"))
	fmt.Printf("DNI válido: %v\n", sunatlib.IsValidDNI("70408005"))
	fmt.Printf("CE válido: %v\n", sunatlib.IsValidCE("001234567"))
}