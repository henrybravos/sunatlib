# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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