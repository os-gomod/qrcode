package payload

import (
	"errors"
	"fmt"
	"strings"
)

type PIDPayload struct {
	PIDType        string
	CreditorName   string
	IBAN           string
	Reference      string
	Amount         string
	Currency       string
	DebtorName     string
	RemittanceInfo string
}

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

func (p *PIDPayload) Validate() error {
	if p.IBAN == "" {
		return errors.New("pid payload: IBAN must not be empty")
	}
	pt := p.pidType()
	if pt != "QRR" && pt != "SCOR" && pt != "NON" && pt != "" {
		return fmt.Errorf("pid payload: unsupported type %q, must be QRR, SCOR, or NON", p.PIDType)
	}
	return nil
}

func (*PIDPayload) Type() string {
	return "pid"
}

func (p *PIDPayload) Size() int {
	encoded, _ := p.Encode()
	return len(encoded)
}

func (p *PIDPayload) pidType() string {
	if p.PIDType == "" {
		return "QRR"
	}
	return p.PIDType
}
