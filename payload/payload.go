package payload

type Payload interface {
	Encode() (string, error)
	Type() string
	Validate() error
	Size() int
}
type BasePayload struct{}

func (*BasePayload) Type() string {
	return "unknown"
}

func (*BasePayload) Size() int {
	return 0
}
