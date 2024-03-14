package mask

import (
	"encoding/json"

	"github.com/showa-93/go-mask"
)

const (
	MaskTypePhone = "phone"
	MaskTypeEmail = "email"
	MaskTypeName  = "name"
)

type Masker struct {
	masker *mask.Masker
}

func New() *Masker {
	masker := mask.NewMasker()

	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
	masker.RegisterMaskStringFunc(mask.MaskTypeFixed, masker.MaskFixedString)
	masker.RegisterMaskStringFunc(MaskTypeEmail, maskEmail)
	masker.RegisterMaskStringFunc(MaskTypePhone, maskPhone)
	masker.RegisterMaskStringFunc(MaskTypeName, maskName)

	return &Masker{masker: masker}
}

// Struct takes a struct value and a list of field names (optional).
// It masks the values of the specified fields in the JSON with a predefined mask.
// The function returns the masked struct as a byte slice or an error if any.
// We encourage you to use the mask tag for readability instead of the optional params.
func (m *Masker) Struct(v any, params ...string) (any, error) {
	for _, p := range params {
		m.masker.RegisterMaskField(p, "filled")
		m.masker.RegisterMaskField(p, "fixed")
	}

	masked, err := m.masker.Mask(v)
	if err != nil {
		return nil, err
	}

	return masked, nil
}

// StructToByte takes a struct value and a list of field names (optional).
// It masks the values of the specified fields in the JSON with a predefined mask.
// The function returns the masked struct as a byte slice or an error if any.
// We encourage you to use the mask tag for readability instead of the optional params.
func (m *Masker) StructToByte(v any, params ...string) ([]byte, error) {
	for _, p := range params {
		m.masker.RegisterMaskField(p, "filled")
		m.masker.RegisterMaskField(p, "fixed")

	}

	masked, err := m.masker.Mask(v)
	if err != nil {
		return nil, err
	}

	mv, err := json.Marshal(masked)
	if err != nil {
		return nil, err
	}

	return mv, nil
}

// MaskJSONBytes takes a JSON byte slice and a list of field names.
// It masks the values of the specified fields in the JSON with a predefined mask.
// The function returns the masked JSON byte slice or an error if any.
// If you have the struct opt for the Struct function instead, and complement it using the mask tag.
func (m *Masker) JSONBytes(data []byte, params ...string) ([]byte, error) {
	var mapValue map[string]any

	if err := json.Unmarshal(data, &mapValue); err != nil {
		return nil, err
	}

	for _, p := range params {
		m.masker.RegisterMaskField(p, "filled")
		m.masker.RegisterMaskField(p, "fixed")
	}

	masked, err := m.masker.Mask(mapValue)
	if err != nil {
		return nil, err
	}

	mv, err := json.Marshal(masked)
	if err != nil {
		return nil, err
	}

	return mv, nil
}
