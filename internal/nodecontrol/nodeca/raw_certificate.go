package nodeca

import "bytes"

func checkAlgorithm(v derElement) error {
	body := v.content
	oid, e := takeElement(&body, 6)
	if e != nil {
		return e
	}
	if _, e = readOID(oid.full); e != nil {
		return e
	}
	if len(body) > 0 {
		param, tail, e := readElement(body)
		if e != nil || len(tail) != 0 {
			return ErrInvalidMaterial
		}
		return checkDERValue(param)
	}
	return nil
}
func checkName(v derElement) error {
	return checkNameSyntax(v, false)
}
func checkNameSyntax(v derElement, generalName bool) error {
	rdns := v.content
	for len(rdns) > 0 {
		rdn, e := takeElement(&rdns, 0x31)
		if e != nil || len(rdn.content) == 0 {
			return ErrInvalidMaterial
		}
		atvs := rdn.content
		var previous []byte
		for len(atvs) > 0 {
			atv, e := takeElement(&atvs, 0x30)
			if e != nil {
				return e
			}
			if previous != nil && derLess(atv.full, previous) {
				return ErrInvalidMaterial
			}
			previous = atv.full
			body := atv.content
			oid, e := takeElement(&body, 6)
			if e != nil {
				return e
			}
			if generalName {
				if !minimalOIDContent(oid.content) {
					return ErrInvalidMaterial
				}
			} else if _, e = readOID(oid.full); e != nil {
				return e
			}
			value, tail, e := readElement(body)
			if e != nil || len(tail) != 0 {
				return ErrInvalidMaterial
			}
			if e = checkDERValue(value); e != nil {
				return e
			}
			if generalName {
				if e = checkNameValue(value); e != nil {
					return e
				}
			}
		}
	}
	return nil
}

// Match the pinned Name parser, not an OID-specific CN or naming policy.
func checkNameValue(v derElement) error {
	switch v.tag {
	case 12, 18, 19, 20, 22, 30:
		return checkUniversalString(v)
	}
	return ErrInvalidMaterial
}
func parseRawCA(der []byte) (rawCA, error) {
	var out rawCA
	outer, e := singleElement(der, 0x30)
	if e != nil {
		return out, e
	}
	parts := outer.content
	if out.tbs, e = takeElement(&parts, 0x30); e != nil {
		return out, e
	}
	if out.outerSignatureAlgorithm, e = takeElement(&parts, 0x30); e != nil {
		return out, e
	}
	if out.signature, e = takeElement(&parts, 3); e != nil || len(parts) != 0 {
		return out, ErrInvalidMaterial
	}
	fields := out.tbs.content
	if len(fields) > 0 && fields[0] == 0xa0 {
		v, e := takeElement(&fields, 0xa0)
		if e != nil {
			return out, e
		}
		version, e := singleElement(v.content, 2)
		if e != nil || !minimalInteger(version.content) {
			return out, ErrInvalidMaterial
		}
	}
	if out.serial, e = takeElement(&fields, 2); e != nil {
		return out, e
	}
	if out.innerSignatureAlgorithm, e = takeElement(&fields, 0x30); e != nil {
		return out, e
	}
	if e = checkAlgorithm(out.innerSignatureAlgorithm); e != nil {
		return out, e
	}
	if e = checkAlgorithm(out.outerSignatureAlgorithm); e != nil {
		return out, e
	}
	if !bytes.Equal(out.innerSignatureAlgorithm.full, out.outerSignatureAlgorithm.full) {
		return out, ErrInvalidMaterial
	}
	if out.issuer, e = takeElement(&fields, 0x30); e != nil {
		return out, e
	}
	if e = checkName(out.issuer); e != nil {
		return out, e
	}
	validity, e := takeElement(&fields, 0x30)
	if e != nil {
		return out, e
	}
	times := validity.content
	out.notBefore, times, e = readElement(times)
	if e != nil {
		return out, e
	}
	out.notAfter, times, e = readElement(times)
	if e != nil || len(times) != 0 || (out.notBefore.tag != 0x17 && out.notBefore.tag != 0x18) || (out.notAfter.tag != 0x17 && out.notAfter.tag != 0x18) {
		return out, ErrInvalidMaterial
	}
	if out.subject, e = takeElement(&fields, 0x30); e != nil {
		return out, e
	}
	if e = checkName(out.subject); e != nil {
		return out, e
	}
	if out.spki, e = takeElement(&fields, 0x30); e != nil {
		return out, e
	}
	if e = checkDERValue(out.spki); e != nil {
		return out, e
	}
	hasUniqueID := false
	for _, tag := range []byte{0x81, 0x82} {
		if len(fields) > 0 && fields[0] == tag {
			id, e := takeElement(&fields, tag)
			if e != nil || !bitStringShape(id.content) {
				return out, ErrInvalidMaterial
			}
			hasUniqueID = true
		}
	}
	if len(fields) > 0 {
		wrapper, e := takeElement(&fields, 0xa3)
		if e != nil || len(fields) != 0 {
			return out, ErrInvalidMaterial
		}
		seq, e := singleElement(wrapper.content, 0x30)
		if e != nil {
			return out, e
		}
		items := seq.content
		for len(items) > 0 {
			item, e := takeElement(&items, 0x30)
			if e != nil {
				return out, e
			}
			body := item.content
			oid, e := takeElement(&body, 6)
			if e != nil {
				return out, e
			}
			id, e := readOID(oid.full)
			if e != nil {
				return out, e
			}
			critical := false
			if len(body) > 0 && body[0] == 1 {
				flag, e := takeElement(&body, 1)
				if e != nil || len(flag.content) != 1 || flag.content[0] != 255 {
					return out, ErrInvalidMaterial
				}
				critical = true
			}
			value, e := takeElement(&body, 4)
			if e != nil || len(body) != 0 {
				return out, ErrInvalidMaterial
			}
			out.extensions = append(out.extensions, rawExtension{id: id, critical: critical, value: value.content})
		}
	}
	if !bitStringShape(out.signature.content) {
		return out, ErrInvalidMaterial
	}
	if len(out.signature.content) == 1 || out.signature.content[0] != 0 || hasUniqueID {
		return out, ErrUnsupportedCAProfile
	}
	return out, nil
}
