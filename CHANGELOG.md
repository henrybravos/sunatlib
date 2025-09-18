# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2025-01-18

### Added
- **Enhanced Ticket Query System for Voided Documents Communications**
  - `QueryVoidedDocumentsTicket()` - Advanced ticket status query with detailed response information
  - `WaitForTicketProcessing()` - Automatic polling with configurable timeout and polling intervals
  - `BatchQueryTickets()` - Query multiple tickets in a single batch operation
  - New `TicketStatusResponse` struct with comprehensive status information and convenience methods

- **Convenience Methods for Ticket Status Management**
  - `IsSuccessful()` - Check if ticket was processed successfully (status code "0")
  - `IsInProgress()` - Check if ticket is still being processed (status code "98")
  - `HasErrors()` - Check if ticket was processed with errors (status code "99")
  - `IsProcessed()` - Check if ticket processing is complete (success or error)
  - `GetTicketStatusDescription()` - Human-readable status descriptions in Spanish
  - `HasApplicationResponse()` - Check if CDR or error data is available
  - `GetApplicationResponse()` - Get CDR or error data as byte array for user handling

- **Enhanced Response Parsing and Error Handling**
  - Improved SOAP fault detection and error message extraction
  - Better HTML entity decoding for Spanish characters (ó, <, >, &)
  - Support for multiple SOAP response formats and namespaces
  - Enhanced error message parsing with proper character encoding

- **Library Usage Examples**
  - `ticket_query_example.go` - Comprehensive examples of all ticket query functionality
  - `library_usage_example.go` - Best practices for library integration and data handling

### Changed
- **Library Philosophy Enhancement - User Control Over Data**
  - Library now strictly follows "provide data, user decides what to do" principle
  - Removed automatic file saving methods - users have complete control over data handling
  - Enhanced flexibility for integration into different systems and workflows
  - Users can now save, process, store, or transmit data according to their specific needs

- **Improved Documentation and API Reference**
  - Updated README with comprehensive ticket query functionality documentation
  - Added library philosophy section emphasizing user control over data handling
  - Enhanced API reference with all new methods and their usage patterns
  - Added real-world integration examples and patterns

### Improved
- **Better Error Handling and Status Management**
  - Enhanced error handling in ticket status queries with detailed error information
  - Improved status code mapping for all SUNAT response scenarios
  - More robust response parsing supporting multiple SOAP formats
  - Better separation of concerns between data retrieval and data handling

- **Code Quality and Maintainability**
  - Cleaner separation between library functionality and user data handling
  - Removed utility functions that imposed file system operations
  - Enhanced modularity for better testing and integration
  - Improved code documentation and inline comments

### Technical Details
- Enhanced SOAP response parsing with support for both `<br:getStatusResponse>` and `getStatusResponse` formats
- Improved base64 decoding for CDR and error content extraction
- Better handling of SUNAT status codes: "0" (success), "98" (processing), "99" (errors)
- Enhanced timeout and polling mechanisms for long-running processes

### Breaking Changes
- **Removed Methods** (replaced with better alternatives):
  - `SaveApplicationResponse()` method removed from `TicketStatusResponse`
  - Users should now use `GetApplicationResponse()` and handle data as needed
- **Philosophy Change**: Library no longer performs file operations automatically
- All other functionality remains fully backward compatible

## [1.3.0] - 2025-09-18

### Added
- **NEW**: Voided Documents Communication (Comunicación de Baja)
  - `SendVoidedDocuments()` method for document cancellation
  - `GetVoidedDocumentsStatus()` method for checking cancellation status
  - `GenerateVoidedDocumentsXML()` for XML generation
  - Asynchronous processing with ticket system
  - Support for multiple documents in single communication
  - Automatic series generation (RA-YYYYMMDD-###)

- **NEW**: Electronic Document Validation (Consulta de Validez)
  - `ValidateInvoice()`, `ValidateReceipt()`, `ValidateCreditNote()`, `ValidateDebitNote()` methods
  - Direct SOAP communication with SUNAT validation service
  - Real-time document status verification
  - Support for all electronic document types

- **NEW**: Beta/Testing Environment Support
  - `NewVoidedDocumentsClientBeta()` for testing voided documents
  - `NewDocumentValidationClientBeta()` for testing document validation
  - Production and Beta endpoints properly separated
  - Test credentials support (MODDATOS/moddatos)
  - Complete testing workflow examples

- **NEW**: Enhanced Validation System
  - RUC format validation with checksum verification
  - Document series and number format validation
  - Document type code validation
  - Special character cleaning for XML safety
  - Complete request validation before SUNAT communication

### Architecture
- **New Files Added**:
  - `voided_documents.go` - Voided documents functionality
  - `document_validation.go` - Document validation functionality
  - `endpoints.go` - Endpoint management and environment selection
  - `utils/validation.go` - Validation utilities

### Technical Details
- Voided documents use UBL VoidedDocuments-1 schema with ISO-8859-1 encoding
- Document validation uses exact SOAP format as required by SUNAT
- Beta endpoints: `https://e-beta.sunat.gob.pe/...`
- Production endpoints: `https://e-factura.sunat.gob.pe/...`
- Date format DD/MM/YYYY for document validation
- CDATA support for company names and void reasons

### Examples Added
- `voided_documents_example.go` - Voided documents communication
- `document_validation_example.go` - Document validation
- `beta_testing_example.go` - Beta environment testing
- `integrated_example.go` - Complete workflow integration

### Breaking Changes
- None - Fully backward compatible
- All existing functionality unchanged

## [1.2.0] - 2025-09-09

### Added
- **NEW**: Independent consultation services (separate from billing)
  - `ConsultationClient` - Independent client for document consultation
  - `NewConsultationClient()` - Full consultation client (RUC + DNI)
  - `NewRUCConsultationClient()` - RUC-only consultation client
  - `NewDNIConsultationClient()` - DNI/CE-only consultation client (free)
- **NEW**: RUC consultation service using DeColecta API
  - `ConsultRUC()` - Basic company information lookup
  - `ConsultRUCFull()` - Complete company information lookup  
  - Requires DeColecta API key (paid service)
- **NEW**: DNI/CE consultation service using EsSalud (free)
  - `ConsultDNI()` - DNI validation and person data retrieval
  - `ConsultCE()` - Carnet de Extranjería validation
  - Always available, no API key required
- **NEW**: Document validation functions
  - `IsValidRUC()` - RUC format validation
  - `IsValidDNI()` - DNI format validation  
  - `IsValidCE()` - CE format validation
- Comprehensive consultation examples in `examples/consultation_example.go`
- Updated documentation with independent consultation service usage

### Architecture
- **Separation of Concerns**: Billing and consultation services are now independent
- **SUNATClient**: Focused exclusively on electronic billing
- **ConsultationClient**: Dedicated to document consultation services
- Modular design allows using only needed services

### Enhanced  
- Separated signing and sending functionality for better control
- `SignXML()` - Sign XML documents independently
- `SendToSUNAT()` - Send pre-signed documents
- Maintained backward compatibility with `SignAndSendInvoice()`

### Services
- **RUC Service**: DeColecta API integration (requires API key)
  - Basic and full company information
  - Real-time SUNAT data
  - Commercial service with usage limits
- **DNI/CE Service**: EsSalud integration (free)
  - Person identity validation
  - No registration required
  - Always available

### Breaking Changes
- None - Fully backward compatible
- All existing billing functionality unchanged

### Requirements
- Go 1.19+
- xmlsec1 system dependency
- Valid SUNAT certificate (for billing)
- Optional: DeColecta API key (for RUC consultation services)

## [1.1.0] - 2025-09-09

### Added
- **NEW**: RUC and DNI consultation services integrated with SUNATClient
- **NEW**: RUC consultation service using DeColecta API
- **NEW**: DNI/CE consultation service using EsSalud (free)
- **NEW**: Document validation functions
- **NEW**: Enhanced SUNATClient constructors with consultation services
- Comprehensive consultation examples
- Updated documentation

### Enhanced
- Separated signing and sending functionality for better control
- `SignXML()` - Sign XML documents independently
- `SendToSUNAT()` - Send pre-signed documents
- Maintained backward compatibility with `SignAndSendInvoice()`

### Breaking Changes
- None - Fully backward compatible

## [1.0.0] - 2025-01-05

### Added
- Initial release of SUNATLib
- XML digital signature functionality using xmlsec1
- Support for PKCS#12 (.pfx) and PEM certificates
- SOAP communication with SUNAT web services
- Automatic ZIP and base64 encoding
- Certificate validation utilities
- CDR (Constancia de Recepción) processing
- Comprehensive examples and documentation
- Support for UBL 2.1 invoice format
- Compatibility with SUNAT Peru electronic billing requirements

### Features
- `NewSUNATClient()` - Create SUNAT client with credentials
- `SetCertificate()` - Configure certificate from PEM files
- `SetCertificateFromPFX()` - Configure certificate from PFX files
- `SignAndSendInvoice()` - Sign XML and send to SUNAT
- `ExtractPEMFromPFX()` - Extract PEM from PFX certificates
- `ValidateCertificate()` - Certificate validation
- `CheckXMLSec1Available()` - Verify xmlsec1 availability

### Requirements
- Go 1.19+
- xmlsec1 system dependency
- Valid SUNAT certificate