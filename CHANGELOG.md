# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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