# Test Data para SUNAT Beta

Archivos XML listos para firmar y enviar al entorno **beta** de SUNAT.

## Endpoint Beta
```
https://e-beta.sunat.gob.pe/ol-ti-itcpfegem-beta/billService
```

## Datos de Prueba (reemplazar con los reales)
| Campo | Valor placeholder | Descripción |
|-------|-------------------|-------------|
| RUC Emisor | `20000000001` | Tu RUC real de prueba |
| RUC Receptor | `20100070970` | RUC genérico de receptor |
| Fechas | `2026-04-27` | Actualizar a fecha actual |

---

## Archivos y Casos

| Archivo | tipAfeIgv | TaxScheme | Cat. 05 | ProfileID | Total |
|---------|-----------|-----------|---------|-----------|-------|
| `F001-00000001_grabado_oneroso.xml` | 10 | 1000 (IGV) | `S` | 0101 | 118.00 PEN |
| `F001-00000002_retiro_bonificacion.xml` | 15 | 9996 (GRA) | `Z` | 0104 | 0.00 PEN (empresa absorbe IGV) |
| `F001-00000003_exonerado.xml` | 20 | 9997 (EXO) | `E` | 0101 | 100.00 PEN |
| `F001-00000004_inafecto.xml` | 30 | 9998 (INA) | `O` | 0101 | 100.00 PEN |
| `F001-00000005_exportacion.xml` | 40 | 9995 (EXP) | `G` | 0200 | 100.00 USD |
| `F001-00000006_mixto.xml` | 10+20+30 | 1000+9997+9998 | mix | 0101 | 218.00 PEN |
| `F001-00000007_isc_igv.xml` | 10+ISC | 2000+1000 | `S` | 0101 | 141.60 PEN |

---

## Cómo firmar y enviar con sunatlib (Go)

```go
client := sunatlib.NewSUNATClient(ruc, usuario, clave, sunatlib.GetBillServiceEndpoint(sunatlib.Beta))
client.SetCertificateFromPFX("certificado.pfx", "password", "/tmp/certs")

xml, _ := os.ReadFile("testdata/F001-00000001_grabado_oneroso.xml")
resp, err := client.SignAndSendInvoice(xml, "01", "F001-00000001")
```

---

## Notas

### Caso Retiro/Bonificación (F001-00000002)
- **ProfileID** debe ser `0104` (Retiro de Bienes), NO `0101`.
- `LineExtensionAmount = 0` — el cliente paga cero.
- `TaxTotal.TaxAmount = 18.00` — la **empresa absorbe** el IGV.
- Dos `AlternativeConditionPrice`: tipo `01` (precio cliente=0) y tipo `02` (valor referencial=100).

### Caso Exportación (F001-00000005)
- **ProfileID** y **InvoiceTypeCode listID** deben ser `0200`.
- Moneda en USD es lo más común pero puede ser PEN.
- Receptor usa `schemeID="0"` (no domiciliado sin documento peruano).

### Caso ISC (F001-00000007)
- ISC **se calcula primero** sobre el valor de venta.
- La base del IGV = valor venta + ISC.
- En la línea hay **dos TaxTotal** separados: uno para ISC (2000) y otro para IGV (1000).
- El ISC usa `TaxTypeCode = EXC`.
