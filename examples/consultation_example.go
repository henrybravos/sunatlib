package main

import (
	"fmt"
	"log"

	"github.com/henrybravos/sunatlib"
)

func main() {
	// Create consultation client (independent from billing)
	// Option 1: Client with both RUC and DNI services
	consultationClient := sunatlib.NewConsultationClient("your_decolecta_api_key")
	
	// Option 2: Only RUC consultation
	// rucClient := sunatlib.NewRUCConsultationClient("your_decolecta_api_key")
	
	// Option 3: Only DNI consultation (free)
	// dniClient := sunatlib.NewDNIConsultationClient()

	// Example 1: Basic RUC consultation
	fmt.Println("=== Consulta Básica de RUC ===")
	rucResult, err := consultationClient.ConsultRUC("20601030013")
	if err != nil {
		log.Printf("Error consultando RUC: %v", err)
	} else if rucResult.Success {
		fmt.Printf("RUC: %s\n", rucResult.Data.RUC)
		fmt.Printf("Razón Social: %s\n", rucResult.Data.RazonSocial)
		fmt.Printf("Estado: %s\n", rucResult.Data.Estado)
		fmt.Printf("Condición: %s\n", rucResult.Data.Condicion)
		fmt.Printf("Dirección: %s\n", rucResult.Data.Direccion)
		fmt.Printf("Distrito: %s, Provincia: %s, Departamento: %s\n", 
			rucResult.Data.Distrito, rucResult.Data.Provincia, rucResult.Data.Departamento)
	} else {
		fmt.Printf("Error: %s\n", rucResult.Message)
	}

	fmt.Println("\n=== Consulta Completa de RUC ===")
	rucFullResult, err := consultationClient.ConsultRUCFull("20601030013")
	if err != nil {
		log.Printf("Error consultando RUC completo: %v", err)
	} else if rucFullResult.Success {
		fmt.Printf("RUC: %s\n", rucFullResult.Data.RUC)
		fmt.Printf("Razón Social: %s\n", rucFullResult.Data.RazonSocial)
		fmt.Printf("Estado: %s\n", rucFullResult.Data.Estado)
		fmt.Printf("Actividad Económica: %s\n", rucFullResult.Data.ActividadEconomica)
		fmt.Printf("Número de Trabajadores: %s\n", rucFullResult.Data.NumeroTrabajadores)
		fmt.Printf("Tipo de Facturación: %s\n", rucFullResult.Data.TipoFacturacion)
		fmt.Printf("Comercio Exterior: %s\n", rucFullResult.Data.ComercioExterior)
	} else {
		fmt.Printf("Error: %s\n", rucFullResult.Message)
	}

	// Example 2: DNI consultation
	fmt.Println("\n=== Consulta de DNI ===")
	dniResult, err := consultationClient.ConsultDNI("12345678")
	if err != nil {
		log.Printf("Error consultando DNI: %v", err)
	} else if dniResult.Success {
		fmt.Printf("DNI: %s\n", dniResult.Data.DNI)
		fmt.Printf("Nombre Completo: %s\n", dniResult.Data.NombreCompleto)
		fmt.Printf("Nombres: %s\n", dniResult.Data.Nombres)
		fmt.Printf("Apellido Paterno: %s\n", dniResult.Data.ApellidoPaterno)
		fmt.Printf("Apellido Materno: %s\n", dniResult.Data.ApellidoMaterno)
	} else {
		fmt.Printf("Error: %s\n", dniResult.Message)
	}

	// Example 3: Carnet de Extranjería consultation
	fmt.Println("\n=== Consulta de Carnet de Extranjería ===")
	ceResult, err := consultationClient.ConsultCE("001234567")
	if err != nil {
		log.Printf("Error consultando CE: %v", err)
	} else if ceResult.Success {
		fmt.Printf("CE: %s\n", ceResult.Data.DNI)
		fmt.Printf("Nombre Completo: %s\n", ceResult.Data.NombreCompleto)
		fmt.Printf("Nombres: %s\n", ceResult.Data.Nombres)
		fmt.Printf("Apellido Paterno: %s\n", ceResult.Data.ApellidoPaterno)
		fmt.Printf("Apellido Materno: %s\n", ceResult.Data.ApellidoMaterno)
	} else {
		fmt.Printf("Error: %s\n", ceResult.Message)
	}

	// Example 4: Validation functions
	fmt.Println("\n=== Validaciones ===")
	testRUC := "20601030013"
	testDNI := "12345678"
	testCE := "001234567"

	fmt.Printf("RUC %s es válido: %v\n", testRUC, sunatlib.IsValidRUC(testRUC))
	fmt.Printf("DNI %s es válido: %v\n", testDNI, sunatlib.IsValidDNI(testDNI))
	fmt.Printf("CE %s es válido: %v\n", testCE, sunatlib.IsValidCE(testCE))

	// Example 5: Using basic client (DNI/CE free, RUC requires API key)
	fmt.Println("\n=== Cliente Básico (Solo DNI/CE gratuitos) ===")
	basicClient := sunatlib.NewSUNATClient(
		"20123456789", 
		"USERNAME", 
		"PASSWORD", 
		"https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService",
	)

	// This should work (DNI service is always available and free)
	dniBasicResult, err := basicClient.ConsultDNI("12345678")
	if err != nil {
		log.Printf("Error consultando DNI con cliente básico: %v", err)
	} else if dniBasicResult.Success {
		fmt.Printf("✅ DNI consultado gratuitamente: %s\n", dniBasicResult.Data.NombreCompleto)
	}

	// This should fail (RUC service requires DeColecta API key)
	_, err = basicClient.ConsultRUC("20601030013")
	if err != nil {
		fmt.Printf("❌ Error esperado - RUC requiere API key: %v\n", err)
	}
}