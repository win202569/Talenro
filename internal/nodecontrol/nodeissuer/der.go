package nodeissuer

import "bytes"

type derElement struct {
	tag           byte
	full, content []byte
}
type signedLeaf struct {
	tbs, issuer, subject, spki                derElement
	innerAlgorithm, outerAlgorithm, signature derElement
}

func readElement(raw []byte) (derElement, []byte, error) {
	var zero derElement
	if len(raw) < 2 || raw[0] == 0 || raw[0]&31 == 31 {
		return zero, nil, ErrInvalidMaterial
	}
	header, length := 2, int(raw[1])
	if raw[1]&128 != 0 {
		n := int(raw[1] & 127)
		if n == 0 || n > len(raw)-2 || raw[2] == 0 {
			return zero, nil, ErrInvalidMaterial
		}
		header += n
		remaining := len(raw) - header
		length = 0
		for _, b := range raw[2:header] {
			if length > remaining/256 {
				return zero, nil, ErrInvalidMaterial
			}
			length = length*256 + int(b)
			if length > remaining {
				return zero, nil, ErrInvalidMaterial
			}
		}
		if length < 128 {
			return zero, nil, ErrInvalidMaterial
		}
	}
	if length > len(raw)-header {
		return zero, nil, ErrInvalidMaterial
	}
	end := header + length
	return derElement{tag: raw[0], full: raw[:end], content: raw[header:end]}, raw[end:], nil
}
func singleElement(raw []byte, tag byte) (derElement, error) {
	v, tail, err := readElement(raw)
	if err != nil || v.tag != tag || len(tail) != 0 {
		return derElement{}, ErrInvalidMaterial
	}
	return v, nil
}
func takeElement(raw *[]byte, tag byte) (derElement, error) {
	if raw == nil {
		return derElement{}, ErrInvalidMaterial
	}
	v, tail, err := readElement(*raw)
	if err != nil || v.tag != tag {
		return derElement{}, ErrInvalidMaterial
	}
	*raw = tail
	return v, nil
}
func minimalInteger(raw []byte) bool {
	return len(raw) > 0 && (len(raw) == 1 || !(raw[0] == 0 && raw[1]&128 == 0 || raw[0] == 255 && raw[1]&128 != 0))
}
func minimalOID(raw []byte) bool {
	if len(raw) == 0 {
		return false
	}
	start := true
	for _, b := range raw {
		if start && b == 0x80 {
			return false
		}
		start = b&128 == 0
	}
	return start
}
func parseLeafEnvelope(der []byte) (signedLeaf, error) {
	var out signedLeaf
	seq, err := singleElement(der, 0x30)
	if err != nil {
		return out, err
	}
	body := seq.content
	tbs, err := takeElement(&body, 0x30)
	if err != nil {
		return out, err
	}
	alg, err := takeElement(&body, 0x30)
	if err != nil {
		return out, err
	}
	sig, err := takeElement(&body, 3)
	if err != nil || len(body) != 0 {
		return out, ErrInvalidMaterial
	}
	fields := tbs.content
	var parts [8]derElement
	for i, tag := range []byte{0xa0, 2, 0x30, 0x30, 0x30, 0x30, 0x30, 0xa3} {
		parts[i], err = takeElement(&fields, tag)
		if err != nil {
			return out, err
		}
	}
	if len(fields) != 0 || !bytes.Equal(parts[0].full, []byte{0xa0, 3, 2, 1, 2}) {
		return out, ErrInvalidMaterial
	}
	return signedLeaf{tbs: tbs, issuer: parts[3], subject: parts[5], spki: parts[6], innerAlgorithm: parts[2], outerAlgorithm: alg, signature: sig}, nil
}
