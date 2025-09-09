# SUNATLib - Librería Go para Firmas Digitales XML SUNAT

Una librería en Go para firmar documentos XML y enviarlos a SUNAT (Superintendencia Nacional de Aduanas y de Administración Tributaria) de Perú.

## Características

- ✅ Firma digital XML compatible con SUNAT usando xmlsec1
- ✅ Soporte para certificados PKCS#12 (.pfx) y PEM
- ✅ Comunicación SOAP con servicios web de SUNAT
- ✅ Manejo automático de ZIP y codificación base64
- ✅ Validación de certificados
- ✅ Procesamiento de respuestas CDR (Constancia de Recepción)
- ✅ **Consulta RUC usando API DeColecta (Pago - Nuevo)**
- ✅ **Consulta DNI/CE usando servicio EsSalud (Gratuito - Nuevo)**

## Requisitos

- Go 1.19 o superior
- xmlsec1 instalado en el sistema

### Instalación de xmlsec1

**Ubuntu/Debian:**
```bash
sudo apt-get install xmlsec1
```

**MacOS:**
```bash
brew install xmlsec1
```

**RHEL/CentOS:**
```bash
sudo yum install xmlsec1
```

## Instalación

```bash
go mod init your-project
go get github.com/henrybravos/sunatlib
```

## Uso Básico

### Opción 1: Proceso completo (todo en uno)
```go
package main

import (
    "fmt"
    "log"
    "os"
    "github.com/henrybravos/sunatlib"
)

func main() {
    // Configure SUNAT client
    client := sunatlib.NewSUNATClient(
        "20123456789",  // Your RUC number
        "MODDATOS",     // SOL username
        "moddatos",     // SOL password
        "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", // Beta endpoint
    )
    defer client.Cleanup()

    // Configure certificate from PFX
    err := client.SetCertificateFromPFX("certificate.pfx", "your_password", "/tmp/certs")
    if err != nil {
        log.Fatal(err)
    }

    // Read invoice XML
    xmlContent, err := os.ReadFile("invoice.xml")
    if err != nil {
        log.Fatal(err)
    }

    // Sign and send (convenience method)
    response, err := client.SignAndSendInvoice(xmlContent, "01", "F001-00000001")
    if err != nil {
        log.Fatal(err)
    }

    // Check result
    if response.Success {
        fmt.Println("✅ Invoice accepted")
        response.SaveApplicationResponse("cdr.zip")
    } else {
        fmt.Printf("❌ Error: %s\n", response.Message)
    }
}
```

### Opción 2: Control separado (Recomendado)
```go
func main() {
    client := sunatlib.NewSUNATClient("20123456789", "MODDATOS", "moddatos", endpoint)
    defer client.Cleanup()
    
    err := client.SetCertificateFromPFX("certificate.pfx", "password", "/tmp/certs")
    if err != nil {
        log.Fatal(err)
    }

    xmlContent, err := os.ReadFile("invoice.xml")
    if err != nil {
        log.Fatal(err)
    }

    // Step 1: Sign XML (you get the signed XML back)
    signedXML, err := client.SignXML(xmlContent)
    if err != nil {
        log.Fatal(err)
    }

    // Optional: Save signed XML for inspection or later use
    os.WriteFile("invoice_signed.xml", signedXML, 0644)

    // Step 2: Send to SUNAT when ready
    response, err := client.SendToSUNAT(signedXML, "01", "F001-00000001")
    if err != nil {
        log.Fatal(err)
    }

    if response.Success {
        fmt.Println("✅ Invoice accepted")
        response.SaveApplicationResponse("cdr.zip")
    }
}
```

## Uso Avanzado

Para mayor control sobre certificados y configuración:

```go
// Extraer certificado PEM de PFX
privateKey, cert, err := utils.ExtractPEMFromPFX("cert.pfx", "password", "/tmp")
if err != nil {
    log.Fatal(err)
}

// Validar certificado
certInfo, err := utils.GetCertificateInfo(cert)
if err != nil {
    log.Fatal(err)
}

// Configurar con archivos PEM
client.SetCertificate(privateKey, cert)
```

## Estructura de Directorios

```
sunatlib/
├── signer/          # Firma XML con xmlsec1
│   └── xmlsigner.go
├── utils/           # Utilidades para certificados
│   └── cert.go
├── examples/        # Ejemplos de uso
│   ├── simple_example.go
│   └── advanced_example.go
├── sunat.go         # Cliente principal SUNAT
└── README.md
```

## Casos de Uso

### 🔧 Firma únicamente (sin envío a SUNAT)
```go
client := sunatlib.NewSUNATClient("", "", "", "") // Sin credenciales SUNAT
client.SetCertificateFromPFX("cert.pfx", "password", "/tmp")

signedXML, err := client.SignXML(xmlContent)
// Usar signedXML para almacenamiento, validación, etc.
```

### 📦 Procesamiento por lotes
```go
documents := []string{"inv1.xml", "inv2.xml", "inv3.xml"}

for _, doc := range documents {
    xmlContent, _ := os.ReadFile(doc)
    signedXML, _ := client.SignXML(xmlContent)
    
    // Decidir cuándo enviar a SUNAT
    if shouldSendNow(doc) {
        client.SendToSUNAT(signedXML, "01", getSeriesNumber(doc))
    } else {
        saveForLater(signedXML, doc)
    }
}
```

### ⏰ Firmar ahora, enviar después
```go
// Fase 1: Firmar durante horario de oficina
signedXML, _ := client.SignXML(xmlContent)
os.WriteFile("signed_invoice.xml", signedXML, 0644)

// Fase 2: Enviar durante ventana de transmisión
signedXML, _ := os.ReadFile("signed_invoice.xml")
response, _ := client.SendToSUNAT(signedXML, "01", "F001-00000001")
```

## Servicios de Consulta (Nuevo)

### Consulta RUC con DeColecta (Requiere API Key - Pago)

```go
// Cliente con servicio RUC habilitado (requiere API key de DeColecta)
client := sunatlib.NewSUNATClientWithRUCService(
    "20123456789",                                    // RUC
    "MODDATOS",                                       // Username  
    "moddatos",                                       // Password
    "https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService", // Endpoint
    "your_decolecta_api_key",                        // DeColecta API Key (pago)
)

// Consulta básica de RUC
rucResult, err := client.ConsultRUC("20601030013")
if err != nil {
    log.Fatal(err)
}

if rucResult.Success {
    fmt.Printf("Razón Social: %s\n", rucResult.Data.RazonSocial)
    fmt.Printf("Estado: %s\n", rucResult.Data.Estado)
    fmt.Printf("Dirección: %s\n", rucResult.Data.Direccion)
}

// Consulta completa de RUC (incluye más detalles)
rucFullResult, err := client.ConsultRUCFull("20601030013")
if err != nil {
    log.Fatal(err)
}

if rucFullResult.Success {
    fmt.Printf("Actividad Económica: %s\n", rucFullResult.Data.ActividadEconomica)
    fmt.Printf("Número de Trabajadores: %s\n", rucFullResult.Data.NumeroTrabajadores)
    fmt.Printf("Tipo de Facturación: %s\n", rucFullResult.Data.TipoFacturacion)
}
```

### Consulta DNI con EsSalud (Gratuito)

```go
// DNI consultation (available in both client types, always free)
// Works with basic client too
basicClient := sunatlib.NewSUNATClient("20123456789", "user", "pass", "endpoint")

dniResult, err := basicClient.ConsultDNI("12345678")
if err != nil {
    log.Fatal(err)
}

if dniResult.Success {
    fmt.Printf("Nombre Completo: %s\n", dniResult.Data.NombreCompleto)
    fmt.Printf("DNI: %s\n", dniResult.Data.DNI)
}

// Carnet de Extranjería consultation
ceResult, err := client.ConsultCE("001234567")
if err != nil {
    log.Fatal(err)
}

if ceResult.Success {
    fmt.Printf("Nombre: %s\n", ceResult.Data.NombreCompleto)
}
```

### Funciones de Validación

```go
// Validar formato de documentos
isValidRUC := sunatlib.IsValidRUC("20601030013")     // true
isValidDNI := sunatlib.IsValidDNI("12345678")        // true  
isValidCE := sunatlib.IsValidCE("001234567")         // true
```

## API Reference

### SUNATClient

#### Métodos

**Constructores:**
- `NewSUNATClient(ruc, username, password, endpoint string) *SUNATClient` - DNI/CE gratuitos incluidos
- `NewSUNATClientWithRUCService(ruc, username, password, endpoint, decolectaAPIKey string) *SUNATClient` - **New!** RUC + DNI/CE

**Configuración de certificados:**
- `SetCertificate(privateKeyPath, certificatePath string) error`
- `SetCertificateFromPFX(pfxPath, password, tempDir string) error`

**Firma y envío a SUNAT:**
- `SignXML(xmlContent []byte) ([]byte, error)`
- `SendToSUNAT(signedXML []byte, documentType, seriesNumber string) (*SUNATResponse, error)`
- `SignAndSendInvoice(xmlContent []byte, documentType, seriesNumber string) (*SUNATResponse, error)`

**Servicios de consulta:** - **New!**
- `ConsultRUC(ruc string) (*RUCBasicResponse, error)`
- `ConsultRUCFull(ruc string) (*RUCFullResponse, error)`
- `ConsultDNI(dni string) (*DNIResponse, error)`
- `ConsultCE(ce string) (*DNIResponse, error)`

**Limpieza:**
- `Cleanup() error`

### SUNATResponse

#### Propiedades

- `Success bool` - Indica si la operación fue exitosa
- `Message string` - Mensaje de respuesta de SUNAT
- `ResponseXML []byte` - XML completo de respuesta
- `ApplicationResponse []byte` - CDR en formato ZIP
- `Error error` - Error si lo hubo

#### Métodos

- `SaveApplicationResponse(outputPath string) error` - Guarda el CDR

### RUCBasicResponse / RUCFullResponse - **New!**

#### Propiedades

- `Success bool` - Indica si la consulta fue exitosa
- `Data *RUCBasicData` / `Data *RUCFullData` - Datos de la empresa consultada
- `Message string` - Mensaje de respuesta

**RUCBasicData campos:**
- `RUC string` - Número de RUC
- `RazonSocial string` - Razón social de la empresa
- `Estado string` - Estado del contribuyente
- `Condicion string` - Condición del contribuyente
- `Direccion string` - Dirección fiscal
- `Distrito string`, `Provincia string`, `Departamento string` - Ubicación

**RUCFullData campos adicionales:**
- `ActividadEconomica string` - Actividad económica principal
- `NumeroTrabajadores string` - Número de trabajadores
- `TipoFacturacion string` - Tipo de sistema de facturación
- `ComercioExterior string` - Si tiene actividad de comercio exterior

### DNIResponse - **New!**

#### Propiedades

- `Success bool` - Indica si la consulta fue exitosa
- `Data *DNIData` - Datos de la persona consultada
- `Message string` - Mensaje de respuesta

**DNIData campos:**
- `DNI string` - Número de documento
- `NombreCompleto string` - Nombre completo
- `Nombres string` - Nombres de la persona
- `ApellidoPaterno string` - Apellido paterno
- `ApellidoMaterno string` - Apellido materno

### Funciones de Validación - **New!**

- `IsValidRUC(ruc string) bool` - Valida formato de RUC
- `IsValidDNI(dni string) bool` - Valida formato de DNI
- `IsValidCE(ce string) bool` - Valida formato de Carnet de Extranjería

### Utils

#### Funciones

- `ExtractPEMFromPFX(pfxPath, password, outputDir string) (privateKeyPath, certPath string, err error)`
- `ValidateCertificate(certPath string) (*x509.Certificate, error)`
- `CheckXMLSec1Available() error`
- `GetCertificateInfo(certPath string) (map[string]string, error)`

## Ejemplos

Ver la carpeta `examples/` para ejemplos completos:

- `simple_example.go` - Uso básico de la librería
- `advanced_example.go` - Manejo avanzado con validación de certificados
- `flexible_usage.go` - Patrones avanzados de uso
- `consultation_example.go` - **Nuevo!** Ejemplos de consulta RUC y DNI

## Limitaciones

- Requiere xmlsec1 instalado en el sistema
- Solo soporta algoritmos RSA-SHA1 (requerimiento SUNAT)
- Diseñado específicamente para documentos UBL 2.1 de SUNAT Perú

## Contribución

Las contribuciones son bienvenidas. Por favor:

1. Fork del proyecto
2. Crear rama para tu feature
3. Commit de cambios
4. Push a la rama
5. Crear Pull Request

## Licencia

MIT License

## Soporte

Para reportar bugs o solicitar features, crear un issue en el repositorio.