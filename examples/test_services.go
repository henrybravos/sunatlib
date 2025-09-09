package main

import (
	"fmt"
	"log"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Test RUC service (requires DeColecta API key)
	fmt.Println("=== Probando servicio RUC (DeColecta - Pago) ===")
	client := sunatlib.NewSUNATClientWithRUCService(
		"20123456789", 
		"MODDATOS", 
		"moddatos", 
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService",
		"sk_10247.COmJdrgJETSqcG5lN503egF10e5UsyKU",
	)
	
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