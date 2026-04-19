package payload

import (
	"fmt"
	"strings"
)

// PIDPayload encodes a Swiss QR-bill payment instruction in the PID format.
// The structured creditor data is encoded as a pipe-separated string after
// the "PID:" prefix. This is used in the Swiss payment infrastructure.
//
// Supported PID types: QRR (QR-referenz, default), SCOR (creditor reference),
// and NON (no structured reference).
//
// Example encoded output:
//
//	PID:QRR|Acme Corp|CH9300762011623852957|RF12345678|150.00|CHF|John Doe|Invoice 2025
type PIDPayload struct {
	// PIDType is the payment instruction type (QRR, SCOR, or NON).
	PIDType string
	// CreditorName is the creditor's name.
	CreditorName string
	// IBAN is the international bank account number.
	IBAN string
	// Reference is the payment reference.
	Reference string
	// Amount is the payment amount.
	Amount string
	// Currency is the three-letter currency code.
	Currency string
	// DebtorName is the debtor's name.
	DebtorName string
	// RemittanceInfo is additional remittance information.
	RemittanceInfo string
}

// Encode returns a PID: string with pipe-separated fields.
// The IBAN has spaces removed. Only non-empty fields are included.
// PIDType defaults to "QRR" if not set.
func (p *PIDPayload) Encode() (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	fields := []string{p.pidType()}
	if p.CreditorName != "" {
		fields = append(fields, p.CreditorName)
	}
	if p.IBAN != "" {
		fields = append(fields, strings.ReplaceAll(p.IBAN, " ", ""))
	}
	if p.Reference != "" {
		fields = append(fields, p.Reference)
	}
	if p.Amount != "" {
		fields = append(fields, p.Amount)
	}
	if p.Currency != "" {
		fields = append(fields, p.Currency)
	}
	if p.DebtorName != "" {
		fields = append(fields, p.DebtorName)
	}
	if p.RemittanceInfo != "" {
		fields = append(fields, p.RemittanceInfo)
	}
	return "PID:" + strings.Join(fields, "|"), nil
}

// Validate checks that the IBAN is non-empty and the PID type (if set)
// is one of QRR, SCOR, or NON.
func (p *PIDPayload) Validate() error {
	if p.IBAN == "" {
		return fmt.Errorf("pid payload: IBAN must not be empty")
	}
	pt := p.pidType()
	if pt != "QRR" && pt != "SCOR" && pt != "NON" && pt != "" {
		return fmt.Errorf("pid payload: unsupported type %q, must be QRR, SCOR, or NON", p.PIDType)
	}
	return nil
}

// Type returns "pid".
func (p *PIDPayload) Type() string {
	return "pid"
}

// Size returns the byte length of the encoded PID string.
func (p *PIDPayload) Size() int {
	encoded, _ := p.Encode() //nolint:errcheck // Size returns 0 on encode error
	return len(encoded)
}

func (p *PIDPayload) pidType() string {
	if p.PIDType == "" {
		return "QRR"
	}
	return p.PIDType
}
