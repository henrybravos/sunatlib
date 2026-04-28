# Release Summary: SUNATLib v1.5.0

## Resumen Ejecutivo
Esta versión se enfoca en la **robustez y confiabilidad** del proceso de emisión de comprobantes electrónicos (CPE). Se ha implementado un nuevo sistema de validación estructural proactiva y se ha refactorizado el motor de firma para eliminar fragilidades ante variaciones de formato XML.

## Nuevas Funcionalidades
1.  **UBL Structural Validator**: Un nuevo componente (`UBLValidator`) que permite validar localmente el cumplimiento de reglas críticas de SUNAT (como el error 3105 y 3024) antes de consumir recursos de red o certificados.
2.  **Beta Integration Suite**: Herramienta de línea de comandos en `cmd/sendbeta` para validación end-to-end contra el gateway oficial de SUNAT Beta, con soporte para 7 escenarios tributarios maestros.
3.  **Regression Testing Framework**: Sistema de pruebas automáticas que utiliza archivos "Golden Masters" para asegurar que los templates de facturación siempre sean válidos.

## Mejoras Técnicas
- **Motor de Firma Whitespace-Agnostic**: Refactorización del `XMLSigner` para inyectar firmas de manera segura independientemente del nivel de indentación o espacios en blanco del XML original.
- **Sustitución Dinámica de RUC**: El cliente de pruebas ahora puede adaptar automáticamente cualquier XML al RUC del certificado configurado.
- **Corrección ISC/IGV**: Implementación de la lógica de cálculo correcta para operaciones mixtas con ISC, evitando discrepancias de bases imponibles en el gateway.

## Estado de Validación (SUNAT Beta)
- ✅ Factura Grabada (Onerosa)
- ✅ Factura Exonerada
- ✅ Factura Inafecta
- ✅ Operación de Exportación
- ✅ Factura Mixta (Gravado + Exonerado + Inafecto)
- ✅ Operación con ISC + IGV
- 🔲 Operación de Retiro/Bonificación (Pendiente ajuste de base imponible)

---
*Release Date: 2026-04-28*
*Maintainer: Antigravity*
