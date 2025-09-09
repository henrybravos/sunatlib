# SUNATLib Examples

This directory contains various examples demonstrating how to use the SUNATLib library.

## Running Examples

Each example is a standalone Go program. To run an individual example:

```bash
# Basic example
go run simple_example.go

# Advanced example with certificate validation
go run advanced_example.go

# Flexible usage patterns
go run flexible_usage.go

# Consultation services (RUC and DNI)
go run consultation_example.go

# Test services functionality
go run test_services.go
```

## Examples Description

- **`simple_example.go`** - Basic usage of the library for signing and sending invoices
- **`advanced_example.go`** - Advanced usage with certificate validation and error handling
- **`flexible_usage.go`** - Demonstrates flexible patterns like sign-only, batch processing, and deferred sending
- **`consultation_example.go`** - Shows how to use RUC and DNI consultation services
- **`test_services.go`** - Simple test of the consultation services

## Requirements

- Valid SUNAT certificate (.pfx file)
- SOL user credentials for SUNAT
- Optional: DeColecta API key for RUC consultation services
- xmlsec1 installed on your system

## Note

These examples use placeholder credentials and data. Replace them with your actual:
- RUC number
- SOL username/password  
- Certificate file paths
- API keys (for consultation services)
- Document data