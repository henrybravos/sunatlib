# SUNATLib - Librería Go para Firmas Digitales XML SUNAT

Una librería en Go para firmar documentos XML y enviarlos a SUNAT (Superintendencia Nacional de Aduanas y de Administración Tributaria) de Perú.

## Características

- ✅ **Firma digital XML robusta** (whitespace-agnostic) compatible con SUNAT usando xmlsec1
- ✅ **Validación Estructural UBL** (Firewall local para errores 3105, 3024) - **Nuevo**
- ✅ **Suite de Testing Beta** con templates "Golden Masters" - **Nuevo**
- ✅ Soporte para certificados PKCS#12 (.pfx) y PEM
- ✅ Comunicación SOAP con servicios web de SUNAT
- ✅ Manejo automático de ZIP y codificación base64
- ✅ Validación de certificados
- ✅ Procesamiento de respuestas CDR (Constancia de Recepción)
- ✅ **Consulta RUC usando servicio directo SUNAT (Gratuito)**
- ✅ **Consulta DNI/CE usando servicio EsSalud (Gratuito)**
- ✅ **Comunicación de Baja (Anulación de Documentos)**
- ✅ **Consulta de Validez de Documentos Electrónicos**
- ✅ **Validación de CPE con credenciales master SUNAT**

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

## Validación Estructural (UBLValidator)

Antes de firmar y enviar un documento, es recomendable pasarlo por el `UBLValidator`. Esto actúa como un "firewall" local que detecta errores estructurales comunes que SUNAT rechazaría (como el error 3105 o el 3024).

```go
v := sunatlib.NewUBLValidator()
err := v.Validate(xmlContent)
if err != nil {
    fmt.Printf("❌ Error estructural: %v\n", err)
    return
}
```

Este validador verifica:
- Presencia de tags obligatorios (`cbc:UBLVersionID`, `cac:TaxTotal`, etc.).
- Consistencia de códigos de tributos (Catálogo 05).
- Reglas de no repetición de totales de impuestos.

## Testing en Ambiente Beta

La librería incluye una herramienta de integración (`cmd/sendbeta`) y una suite de archivos maestros (`testdata/`) para realizar pruebas de regresión contra el gateway oficial de SUNAT.

Para ejecutar la suite de pruebas:
```bash
go run ./cmd/sendbeta/
```

Esto procesará automáticamente 7 escenarios impositivos (Grabado, Exonerado, Inafecto, Exportación, Mixto, ISC+IGV y Retiro), firmándolos con tu certificado de prueba y reportando la respuesta de SUNAT.

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

## Servicios de Consulta (Independientes de Facturación)

### Cliente de Consulta Completo

```go
// Cliente independiente con ambos servicios (RUC + DNI)
consultationClient := sunatlib.NewConsultationClient("")

// O clientes específicos:
rucClient := sunatlib.NewRUCConsultationClient("")  // Solo RUC (Gratuito)
dniClient := sunatlib.NewDNIConsultationClient()     // Solo DNI/CE (Gratuito)

// Consulta básica de RUC
rucResult, err := consultationClient.ConsultRUC("20601030013")
if err != nil {
    log.Fatal(err)
}

if rucResult.Success {
    fmt.Printf("Razón Social: %s\n", rucResult.Data.RazonSocial)
    fmt.Printf("Estado: %s\n", rucResult.Data.Estado)
    fmt.Printf("Dirección: %s\n", rucResult.Data.Direccion)
}

// Consulta completa de RUC (incluye más detalles)
rucFullResult, err := consultationClient.ConsultRUCFull("20601030013")
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
// DNI consultation (always free, independent from billing)
dniClient := sunatlib.NewDNIConsultationClient()

dniResult, err := dniClient.ConsultDNI("12345678")
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

## Comunicación de Baja (Anulación de Documentos) - **Nuevo!**

### Envío de Comunicación de Baja

```go
// Crear cliente para comunicaciones de baja (PRODUCCIÓN)
client := sunatlib.NewVoidedDocumentsClient("20123456789", "MODDATOS", "moddatos")
defer client.Cleanup()

// Para pruebas, usar cliente BETA:
// client := sunatlib.NewVoidedDocumentsClientBeta("20123456789", "MODDATOS", "moddatos")

// Configurar certificado
err := client.SetCertificateFromPFX("certificate.pfx", "password", "/tmp/certs")
if err != nil {
    log.Fatal(err)
}

// Crear solicitud de baja
now := time.Now()
referenceDate := now.AddDate(0, 0, -1) // Documentos de ayer

request := &sunatlib.VoidedDocumentsRequest{
    RUC:           "20123456789",
    CompanyName:   "MI EMPRESA S.A.C.",
    SeriesNumber:  sunatlib.GenerateVoidedDocumentsSeries(referenceDate, 1), // RA-YYYYMMDD-001
    IssueDate:     now,
    ReferenceDate: referenceDate,
    Description:   "Comunicación de baja de documentos",
    Documents: []sunatlib.VoidedDocument{
        {
            DocumentTypeCode: "01",     // Factura
            DocumentSeries:   "F001",   // Serie
            DocumentNumber:   "000123", // Número
            VoidedReason:     "Error en datos del cliente",
        },
        {
            DocumentTypeCode: "03",     // Boleta
            DocumentSeries:   "B001",   // Serie
            DocumentNumber:   "000456", // Número
            VoidedReason:     "Duplicado por error del sistema",
        },
    },
}

// Enviar comunicación de baja
response, err := client.SendVoidedDocuments(request)
if err != nil {
    log.Fatal(err)
}

if response.Success {
    fmt.Printf("✅ Comunicación enviada. Ticket: %s\n", response.Ticket)

    // Consultar estado usando el ticket (método básico)
    statusResponse, err := client.GetVoidedDocumentsStatus(response.Ticket)
    if err == nil && statusResponse.Success {
        statusResponse.SaveApplicationResponse("baja_cdr.zip")
    }

    // O usar el método mejorado para consulta de tickets
    ticketResponse, err := client.QueryVoidedDocumentsTicket(response.Ticket)
    if err == nil {
        if ticketResponse.IsSuccessful() {
            fmt.Println("✅ Procesado exitosamente")
            ticketResponse.SaveApplicationResponse("baja_cdr.zip")
        } else if ticketResponse.IsInProgress() {
            fmt.Println("⏳ Aún en proceso...")
        } else if ticketResponse.HasErrors() {
            fmt.Println("❌ Procesado con errores")
        }
    }
}
```

### Consulta Avanzada de Tickets de Comunicación de Baja - **Nuevo!**

**🎯 Filosofía de la Librería:** Esta librería proporciona los datos, el usuario decide qué hacer con ellos. No fuerza dónde o cómo guardar archivos.

```go
// Consulta individual de ticket con información detallada
ticketResponse, err := client.QueryVoidedDocumentsTicket("12345678901234567890")
if err != nil {
    log.Fatal(err)
}

// Verificar estado usando métodos de conveniencia
if ticketResponse.IsSuccessful() {
    fmt.Println("✅ Comunicación procesada exitosamente")

    // La librería proporciona los datos, el usuario decide qué hacer
    if ticketResponse.HasApplicationResponse() {
        cdrData := ticketResponse.GetApplicationResponse()

        // El usuario puede:
        // 1. Guardar donde desee
        os.WriteFile("mi_directorio/cdr.zip", cdrData, 0644)

        // 2. Procesar directamente
        fmt.Printf("CDR size: %d bytes\n", len(cdrData))

        // 3. Enviar por email, subir a la nube, guardar en BD, etc.
        // sendEmail(cdrData)
        // uploadToCloud(cdrData)
        // saveToDatabase(ticketResponse.Ticket, cdrData)
    }
} else if ticketResponse.IsInProgress() {
    fmt.Println("⏳ Comunicación en proceso...")
} else if ticketResponse.HasErrors() {
    fmt.Println("❌ Comunicación procesada con errores")

    // Obtener detalles del error para análisis del usuario
    if ticketResponse.HasApplicationResponse() {
        errorData := ticketResponse.GetApplicationResponse()
        // Usuario decide cómo manejar los errores
        analyzeErrors(errorData)
        logToSystem(ticketResponse.Ticket, errorData)
    }
}

// Esperar procesamiento con timeout
finalResponse, err := client.WaitForTicketProcessing(
    "12345678901234567890",
    10*time.Minute, // Tiempo máximo de espera
    30*time.Second, // Intervalo de consulta
)

// Consulta múltiple de tickets
tickets := []string{"ticket1", "ticket2", "ticket3"}
responses, err := client.BatchQueryTickets(tickets)
if err == nil {
    for _, response := range responses {
        fmt.Printf("Ticket %s: %s\n", response.Ticket, response.GetTicketStatusDescription())
    }
}
```

## Consulta de Validez de Documentos Electrónicos - **Nuevo!**

### Validación de Documentos con SOAP SUNAT

```go
// Crear cliente de validación con credenciales SOL (PRODUCCIÓN)
client := sunatlib.NewDocumentValidationClient(
    "20123456789", // RUC
    "MODDATOS",    // Usuario SOL
    "moddatos",    // Clave SOL
)

// Para pruebas, usar cliente BETA:
// client := sunatlib.NewDocumentValidationClientBeta("20123456789", "MODDATOS", "moddatos")

// Validar una factura
response, err := client.ValidateInvoice(
    "20123456789", // RUC emisor
    "F001",        // Serie
    "000123",      // Número
    "15/01/2025",  // Fecha emisión (DD/MM/YYYY)
    "118.00",      // Importe total
)

if err != nil {
    log.Fatal(err)
}

if response.Success {
    fmt.Printf("✅ Documento válido: %t\n", response.IsDocumentValid())
    fmt.Printf("📄 Estado: %s\n", response.GetStatusDescription())

    if response.IsValid {
        fmt.Println("🎯 Documento VÁLIDO en SUNAT")
    } else {
        fmt.Println("❌ Documento INVÁLIDO")
    }
} else {
    fmt.Printf("❌ Error: %s\n", response.GetErrorMessage())
}

// Otros métodos de validación disponibles:
receiptResp, _ := client.ValidateReceipt("20123456789", "B001", "000456", "15/01/2025", "59.00")
creditNoteResp, _ := client.ValidateCreditNote("20123456789", "FC01", "000001", "15/01/2025", "23.60")
debitNoteResp, _ := client.ValidateDebitNote("20123456789", "FD01", "000001", "15/01/2025", "15.00")

// Consulta básica de estado (sin fecha ni importe)
statusResp, _ := client.CheckDocumentStatus("20123456789", "01", "F001", "000789")
```

## Validación CPE con Credenciales Master - **Nuevo!**

### Validación usando credenciales master SUNAT

```go
// Crear cliente de validación con credenciales master
validator := sunatlib.NewValidationClient(
    masterRUC,      // Master RUC (parámetro)
    masterUsername, // Master username (parámetro)
    masterPassword, // Master password (parámetro)
)

// Validar factura
invoiceResult, err := validator.ValidateInvoice(
    "20123456789",     // RUC emisor
    "F001",            // Serie
    "00000001",        // Número
    "2024-01-15",      // Fecha emisión (YYYY-MM-DD)
    1250.50,           // Importe total
)

if err != nil {
    log.Printf("Error: %v", err)
} else {
    fmt.Printf("Estado: %s\n", invoiceResult.State)         // VALIDO, NO_INFORMADO, ANULADO, RECHAZADO
    fmt.Printf("Es válido: %t\n", invoiceResult.IsValid)    // true/false
    fmt.Printf("Mensaje: %s\n", invoiceResult.StatusMessage)
}

// Validar boleta
receiptResult, err := validator.ValidateReceipt(
    "20123456789",     // RUC emisor
    "B001",            // Serie
    "00000001",        // Número
    "2024-01-15",      // Fecha emisión (YYYY-MM-DD)
    85.50,             // Importe total
)

// Validación personalizada con parámetros completos
customParams := &sunatlib.ValidationParams{
    IssuerRUC:           "20123456789",
    DocumentType:        "01", // 01=Factura, 03=Boleta
    SeriesNumber:        "F001",
    DocumentNumber:      "00000002",
    RecipientDocType:    "6", // 6=RUC, 1=DNI, 4=CE
    RecipientDocNumber:  "20987654321",
    IssueDate:           "2024-01-15", // YYYY-MM-DD
    TotalAmount:         2500.75,
    AuthorizationNumber: "", // Generalmente vacío
}

result, err := validator.ValidateDocument(customParams)
```

## Endpoints y Ambientes - **Nuevo!**

### Endpoints de Producción vs Beta/Pruebas

La librería incluye soporte completo para endpoints tanto de producción como de pruebas (BETA):

```go
// ENDPOINTS DE PRODUCCIÓN
// Facturación electrónica
client := sunatlib.NewVoidedDocumentsClient("20123456789", "USUARIO", "PASSWORD")

// Validación de documentos
validationClient := sunatlib.NewDocumentValidationClient("20123456789", "USUARIO", "PASSWORD")

// ENDPOINTS DE PRUEBAS (BETA)
// Facturación electrónica (para testing)
betaClient := sunatlib.NewVoidedDocumentsClientBeta("20123456789", "MODDATOS", "moddatos")

// Validación de documentos (para testing)
betaValidationClient := sunatlib.NewDocumentValidationClientBeta("20123456789", "MODDATOS", "moddatos")
```

### Endpoints Disponibles

**Producción:**

- Facturación: `https://e-factura.sunat.gob.pe/ol-ti-itcpfegem/billService`
- Validación: `https://e-factura.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService`

**Beta/Pruebas:**

- Facturación: `https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService`
- Validación: `https://e-beta.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService`

### Credenciales de Prueba

Para el ambiente BETA, usar las credenciales estándar de SUNAT:

- **Usuario:** MODDATOS
- **Contraseña:** moddatos

### Flujo Recomendado de Desarrollo

1. **Desarrollo:** Usar endpoints BETA con credenciales de prueba
2. **Testing:** Validar toda la funcionalidad en BETA
3. **Producción:** Cambiar a endpoints de producción con credenciales reales

## API Reference

### SUNATClient

#### Métodos

**Constructores:**

- `NewSUNATClient(ruc, username, password, endpoint string) *SUNATClient` - Cliente de facturación electrónica

**Constructores de Consulta:** - **New!**

- `NewConsultationClient(apiKey string) *ConsultationClient` - RUC + DNI/CE (Gratuito)
- `NewRUCConsultationClient(apiKey string) *ConsultationClient` - Solo RUC (Gratuito)
- `NewDNIConsultationClient() *ConsultationClient` - Solo DNI/CE (Gratuito)

**Constructores de Baja y Validación:** - **New!**

- `NewVoidedDocumentsClient(ruc, username, password string) *SUNATClient` - Comunicaciones de baja (PRODUCCIÓN)
- `NewVoidedDocumentsClientBeta(ruc, username, password string) *SUNATClient` - Comunicaciones de baja (BETA/Pruebas)
- `NewDocumentValidationClient(ruc, username, password string) *DocumentValidationClient` - Validación de documentos (PRODUCCIÓN)
- `NewDocumentValidationClientBeta(ruc, username, password string) *DocumentValidationClient` - Validación de documentos (BETA/Pruebas)

**Configuración de certificados:**

- `SetCertificate(privateKeyPath, certificatePath string) error`
- `SetCertificateFromPFX(pfxPath, password, tempDir string) error`

**Firma y envío a SUNAT:**

- `SignXML(xmlContent []byte) ([]byte, error)`
- `SendToSUNAT(signedXML []byte, documentType, seriesNumber string) (*SUNATResponse, error)`
- `SignAndSendInvoice(xmlContent []byte, documentType, seriesNumber string) (*SUNATResponse, error)`

**Comunicaciones de Baja:** - **New!**

- `SendVoidedDocuments(request *VoidedDocumentsRequest) (*VoidedDocumentsResponse, error)`
- `GetVoidedDocumentsStatus(ticket string) (*SUNATResponse, error)`
- `QueryVoidedDocumentsTicket(ticket string) (*TicketStatusResponse, error)` - **Nuevo!**
- `WaitForTicketProcessing(ticket string, maxWaitTime, pollInterval time.Duration) (*TicketStatusResponse, error)` - **Nuevo!**
- `BatchQueryTickets(tickets []string) ([]*TicketStatusResponse, error)` - **Nuevo!**
- `GenerateVoidedDocumentsXML(request *VoidedDocumentsRequest) ([]byte, error)`
- `GenerateVoidedDocumentsSeries(referenceDate time.Time, sequential int) string`

### ConsultationClient - **New!**

**Métodos de consulta:**

- `ConsultRUC(ruc string) (*RUCBasicResponse, error)` - Consulta básica RUC
- `ConsultRUCFull(ruc string) (*RUCFullResponse, error)` - Consulta completa RUC
- `ConsultDNI(dni string) (*DNIResponse, error)` - Consulta DNI
- `ConsultCE(ce string) (*DNIResponse, error)` - Consulta CE

**Limpieza:**

- `Cleanup() error`

### DocumentValidationClient - **New!**

**Métodos de validación:**

- `ValidateDocument(req *ValidationRequest) (*ValidationResponse, error)` - Validación genérica
- `ValidateInvoice(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error)` - Validar factura
- `ValidateReceipt(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error)` - Validar boleta
- `ValidateCreditNote(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error)` - Validar nota de crédito
- `ValidateDebitNote(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error)` - Validar nota de débito
- `CheckDocumentStatus(ruc, documentType, series, number string) (*ValidationResponse, error)` - Consulta básica de estado

### VoidedDocumentsRequest - **New!**

**Estructura para comunicaciones de baja:**

- `RUC string` - RUC de la empresa
- `CompanyName string` - Razón social de la empresa
- `SeriesNumber string` - Número de serie de la comunicación (RA-YYYYMMDD-###)
- `IssueDate time.Time` - Fecha de emisión de la comunicación
- `ReferenceDate time.Time` - Fecha de referencia (fecha de los documentos a anular)
- `Documents []VoidedDocument` - Lista de documentos a anular
- `Description string` - Descripción de la comunicación

### VoidedDocument - **New!**

**Estructura para documentos a anular:**

- `DocumentTypeCode string` - Código de tipo de documento (01=Factura, 03=Boleta, etc.)
- `DocumentSeries string` - Serie del documento (F001, B001, etc.)
- `DocumentNumber string` - Número correlativo del documento
- `VoidedReason string` - Motivo de la anulación

### SUNATResponse

#### Propiedades

- `Success bool` - Indica si la operación fue exitosa
- `Message string` - Mensaje de respuesta de SUNAT
- `ResponseXML []byte` - XML completo de respuesta
- `ApplicationResponse []byte` - CDR en formato ZIP
- `Error error` - Error si lo hubo

#### Métodos

- `SaveApplicationResponse(outputPath string) error` - Guarda el CDR

### TicketStatusResponse - **New!**

#### Propiedades

- `Success bool` - Indica si la consulta fue exitosa
- `Message string` - Mensaje de respuesta
- `Ticket string` - Número de ticket consultado
- `StatusCode string` - Código de estado de SUNAT (0=Exitoso, 98=En proceso, 99=Con errores)
- `StatusDescription string` - Descripción del estado
- `ProcessDate time.Time` - Fecha de procesamiento
- `ResponseXML []byte` - XML completo de respuesta
- `ApplicationResponse []byte` - CDR en formato ZIP si está disponible
- `Error error` - Error si lo hubo

#### Métodos

- `GetTicketStatusDescription() string` - Descripción legible del estado del ticket
- `IsProcessed() bool` - True si el ticket fue procesado (exitoso o con errores)
- `IsSuccessful() bool` - True si el ticket fue procesado exitosamente
- `IsInProgress() bool` - True si el ticket aún está en proceso
- `HasErrors() bool` - True si el ticket fue procesado con errores
- `GetApplicationResponse() []byte` - Obtiene los datos del CDR/respuesta
- `HasApplicationResponse() bool` - True si hay datos de respuesta disponibles

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
- `voided_documents_example.go` - **Nuevo!** Ejemplos de comunicaciones de baja
- `document_validation_example.go` - **Nuevo!** Ejemplos de validación de documentos
- `beta_testing_example.go` - **Nuevo!** Ejemplos de testing con endpoints BETA
- `integrated_example.go` - **Nuevo!** Ejemplo completo integrando todas las funcionalidades
- `ticket_query_example.go` - **Nuevo!** Ejemplos de consulta avanzada de tickets de comunicación de baja

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
