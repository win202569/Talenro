package nodeca

import (
	"bytes"
	"crypto/x509"
	"encoding/asn1"
	"time"
)

type extensionSummary struct {
	SubjectKeyIdentifier, AuthorityKeyIdentifier KeyIdentifier
	HasPathLen                                   bool
	PathLen                                      uint8
	CRLSignAllowed                               bool
}

func checkCAExtensions(cert *x509.Certificate, raw []rawExtension) (extensionSummary, error) {
	var out extensionSummary
	if cert == nil {
		return out, ErrInvalidMaterial
	}
	// Fixed slots, not map iteration, define inventory and value error ordering.
	var slots [4]*rawExtension
	if len(raw) != 4 {
		return out, ErrUnsupportedCAProfile
	}
	for i := range raw {
		index := -1
		for j, last := range []int{19, 15, 14, 35} {
			if raw[i].id.Equal(asn1.ObjectIdentifier{2, 5, 29, last}) {
				index = j
				break
			}
		}
		if index < 0 || slots[index] != nil {
			return out, ErrUnsupportedCAProfile
		}
		slots[index] = &raw[i]
	}
	for i, v := range slots {
		if v == nil || v.critical != (i < 2) {
			return out, ErrUnsupportedCAProfile
		}
	}
	if len(cert.Extensions) != 4 {
		return out, ErrInvalidMaterial
	}
	for _, v := range slots {
		found := false
		for _, parsed := range cert.Extensions {
			if parsed.Id.Equal(v.id) {
				if found || parsed.Critical != v.critical || !bytes.Equal(parsed.Value, v.value) {
					return out, ErrInvalidMaterial
				}
				found = true
			}
		}
		if !found {
			return out, ErrInvalidMaterial
		}
	}
	has, path, e := checkBasicConstraints(slots[0].value)
	if e != nil {
		return out, e
	}
	expectedPath := -1
	if has {
		expectedPath = int(path)
	}
	if !cert.BasicConstraintsValid || !cert.IsCA || cert.MaxPathLen != expectedPath || cert.MaxPathLenZero != (has && path == 0) {
		return out, ErrInvalidMaterial
	}
	crl, e := checkKeyUsage(slots[1].value)
	if e != nil {
		return out, e
	}
	usage := x509.KeyUsageCertSign
	if crl {
		usage |= x509.KeyUsageCRLSign
	}
	if cert.KeyUsage != usage {
		return out, ErrInvalidMaterial
	}
	ski, e := singleElement(slots[2].value, 4)
	if e != nil {
		return out, e
	}
	if len(ski.content) != 32 {
		return out, ErrUnsupportedCAProfile
	}
	if !bytes.Equal(ski.content, cert.SubjectKeyId) {
		return out, ErrInvalidMaterial
	}
	aki, e := singleElement(slots[3].value, 0x30)
	if e != nil {
		return out, e
	}
	var ids []derElement
	for rest := aki.content; len(rest) > 0; {
		v, tail, e := readElement(rest)
		if e != nil {
			return out, e
		}
		// Even excluded fields must have complete nested structural encoding
		// before an interpretable out-of-profile AKI is classified below.
		if e := checkDERValue(v); e != nil {
			return out, e
		}
		// [2] is a known IMPLICIT INTEGER, not opaque parameter content.
		if v.tag == 0x82 && !minimalInteger(v.content) {
			return out, ErrInvalidMaterial
		}
		if v.tag == 0xa1 {
			// Consume known GeneralName encodings before rejecting this profile.
			for names := v.content; len(names) > 0; {
				name, tail, e := readElement(names)
				if e != nil {
					return out, e
				}
				if e = checkGeneralName(name); e != nil {
					return out, e
				}
				names = tail
			}
		}
		ids = append(ids, v)
		rest = tail
	}
	if len(ids) != 1 || ids[0].tag != 0x80 || len(ids[0].content) != 32 {
		return out, ErrUnsupportedCAProfile
	}
	if !bytes.Equal(ids[0].content, cert.AuthorityKeyId) {
		return out, ErrInvalidMaterial
	}
	copy(out.SubjectKeyIdentifier[:], ski.content)
	copy(out.AuthorityKeyIdentifier[:], ids[0].content)
	out.HasPathLen = has
	out.PathLen = path
	out.CRLSignAllowed = crl
	return out, nil
}

func checkGeneralName(v derElement) error {
	switch v.tag {
	case 0x81, 0x82, 0x86:
		return checkImplicitString(22, v.content)
	case 0x87:
		return nil // Encoded OCTET STRING; no address/naming policy in this rejected profile.
	case 0x88:
		if !minimalOIDContent(v.content) {
			return ErrInvalidMaterial
		}
		return nil
	case 0xa0:
		body := v.content
		oid, e := takeElement(&body, 6)
		if e != nil || !minimalOIDContent(oid.content) {
			return ErrInvalidMaterial
		}
		value, e := takeElement(&body, 0xa0)
		if e != nil || len(body) != 0 {
			return ErrInvalidMaterial
		}
		return checkExplicitGeneralValue(value)
	case 0xa4:
		name, e := singleElement(v.content, 0x30)
		if e != nil {
			return e
		}
		return checkNameSyntax(name, true)
	case 0xa5:
		body := v.content
		if len(body) > 0 && body[0] == 0xa0 {
			assigner, e := takeElement(&body, 0xa0)
			if e != nil {
				return e
			}
			if e = checkDirectoryString(assigner); e != nil {
				return e
			}
		}
		party, e := takeElement(&body, 0xa1)
		if e != nil || len(body) != 0 {
			return ErrInvalidMaterial
		}
		return checkDirectoryString(party)
	case 0xa3:
		return checkORAddress(v.content)
	}
	return ErrInvalidMaterial
}
func checkExplicitGeneralValue(v derElement) error {
	value, tail, e := readElement(v.content)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	return checkGeneralValue(value)
}
func checkDirectoryString(v derElement) error {
	value, tail, e := readElement(v.content)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	switch value.tag {
	case 12, 19, 20, 28, 30:
		return checkGeneralValue(value)
	}
	return ErrInvalidMaterial
}

// This walk validates known string encodings inside rejected GeneralName values.
// It neither resolves a vendor-defined schema nor interprets any naming authority.
func checkGeneralValue(v derElement) error {
	if e := checkDERValue(v); e != nil {
		return e
	}
	pending := []derElement{v}
	for len(pending) > 0 {
		v = pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if e := checkGeneralNode(v); e != nil {
			return e
		}
		if v.tag&32 != 0 {
			for rest := v.content; len(rest) > 0; {
				child, tail, e := readElement(rest)
				if e != nil {
					return e
				}
				pending = append(pending, child)
				rest = tail
			}
		}
	}
	return nil
}
func checkGeneralNode(v derElement) error {
	switch v.tag {
	case 12, 18, 19, 20, 22, 26, 27, 28, 30:
		return checkUniversalString(v)
	case 13:
		if !minimalOIDContent(v.content) {
			return ErrInvalidMaterial
		}
	case 23, 24:
		return checkGeneralTime(v)
	case 14:
		return checkTIME(v.content)
	case 9:
		return checkGeneralReal(v.content)
	case 0x28:
		return checkExternalValue(v.content)
	case 0x2b, 0x3d:
		return checkEmbeddedValue(v.content)
	}
	return nil
}

// Universal time encoding only; no current time or certificate lifetime policy.
// Fractions remain exact input bytes and are not constrained to Go nanoseconds.
func checkGeneralTime(v derElement) error {
	base := 12
	if v.tag == 24 {
		base = 14
	}
	b := v.content
	if len(b) < base+1 || b[len(b)-1] != 'Z' {
		return ErrInvalidMaterial
	}
	for _, c := range b[:base] {
		if c < '0' || c > '9' {
			return ErrInvalidMaterial
		}
	}
	if len(b) > base+1 {
		if v.tag != 24 || len(b) < base+3 || b[base] != '.' || b[len(b)-2] == '0' {
			return ErrInvalidMaterial
		}
		for _, c := range b[base+1 : len(b)-1] {
			if c < '0' || c > '9' {
				return ErrInvalidMaterial
			}
		}
	}
	// The bounded temporary checks the calendar independently of fraction precision;
	// it is never substituted for the original extension or its opaque value.
	raw := make([]byte, base+3)
	defer clear(raw)
	raw[0], raw[1] = v.tag, byte(base+1)
	copy(raw[2:], b[:base])
	raw[len(raw)-1] = 'Z'
	// Generic GeneralizedTime can represent a leap second that Go's time.Time
	// cannot. Validate the same calendar minute with 59 only in this cleared
	// temporary; do not rewrite the input or authenticate a leap-second schedule.
	// UTCTime retains its distinct 00..59 representation limit.
	if v.tag == 24 && b[base-2] == '6' && b[base-1] == '0' {
		raw[base], raw[base+1] = '5', '9'
	}
	var parsed time.Time
	tail, e := asn1.Unmarshal(raw, &parsed)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	return nil
}

// Validate TIME's value-notation envelope and DER canonical transformations.
// ASCII tstring characters are a strict subset of UTF-8; arbitrary bytes cannot
// masquerade as an opaque time. No numeric year/fraction/duration/recurrence cap,
// wall clock, leap-second schedule, or vendor-defined TIME subtype is inferred.
func checkTIME(b []byte) error {
	if len(b) == 0 {
		return ErrInvalidMaterial
	}
	for _, c := range b {
		if c >= '0' && c <= '9' {
			continue
		}
		switch c {
		case '+', '-', ':', '.', '/', 'C', 'D', 'H', 'M', 'R', 'P', 'S', 'T', 'W', 'Y', 'Z':
		default:
			return ErrInvalidMaterial
		}
	}
	recurring := b[0] == 'R'
	if recurring {
		i := 1
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
		}
		if i >= len(b) || b[i] != '/' || i > 2 && b[1] == '0' {
			return ErrInvalidMaterial
		}
		b = b[i+1:]
		if len(b) == 0 {
			return ErrInvalidMaterial
		}
	}
	slash := -1
	for i, c := range b {
		if c == '/' {
			if slash >= 0 {
				return ErrInvalidMaterial
			}
			slash = i
		}
	}
	if slash < 0 {
		if b[0] == 'P' {
			return checkTimeDuration(b)
		}
		_, _, ok := checkTimePoint(b)
		if !ok || recurring {
			return ErrInvalidMaterial
		}
		return nil
	}
	left, right := b[:slash], b[slash+1:]
	if len(left) == 0 || len(right) == 0 {
		return ErrInvalidMaterial
	}
	if left[0] == 'P' || right[0] == 'P' {
		if left[0] == 'P' && right[0] == 'P' {
			return ErrInvalidMaterial
		}
		duration, point := left, right
		if right[0] == 'P' {
			duration, point = right, left
		}
		_, _, ok := checkTimePoint(point)
		if !ok {
			return ErrInvalidMaterial
		}
		return checkTimeDuration(duration)
	}
	a, az, aok := checkTimePoint(left)
	c, cz, cok := checkTimePoint(right)
	if !aok || !cok {
		return ErrInvalidMaterial
	}
	// A repeated end-point time difference is omitted in canonical DER.
	// A bare end point may inherit that difference, but not a UTC Z suffix.
	if a[6] == 2 && c[6] == 0 {
		c[6] = 2
	} else if a[6] == 2 && c[6] == 2 && bytes.Equal(az, cz) {
		return ErrInvalidMaterial
	}
	if a != c {
		return ErrInvalidMaterial
	}
	return nil
}
func checkTimeDuration(b []byte) error {
	if len(b) < 3 || b[0] != 'P' {
		return ErrInvalidMaterial
	}
	i, rank := 1, 0
	inTime := false
	for i < len(b) {
		if b[i] == 'T' {
			if inTime {
				return ErrInvalidMaterial
			}
			inTime = true
			i++
		}
		start := i
		zero := true
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			zero = zero && b[i] == '0'
			i++
		}
		if i == start || i-start > 1 && b[start] == '0' {
			return ErrInvalidMaterial
		}
		fraction := false
		if i < len(b) && b[i] == '.' {
			fraction = true
			i++
			at := i
			for i < len(b) && b[i] >= '0' && b[i] <= '9' {
				zero = zero && b[i] == '0'
				i++
			}
			if at == i {
				return ErrInvalidMaterial
			}
		}
		if i >= len(b) {
			return ErrInvalidMaterial
		}
		next := 0
		switch b[i] {
		case 'Y':
			if !inTime {
				next = 1
			}
		case 'M':
			next = 2
			if inTime {
				next = 5
			}
		case 'D':
			if !inTime {
				next = 3
			}
		case 'H':
			if inTime {
				next = 4
			}
		case 'S':
			if inTime {
				next = 6
			}
		case 'W':
			if !inTime && rank == 0 && i+1 == len(b) {
				next = 7
			}
		}
		i++
		if next <= rank || next == 0 || (fraction || zero) && i != len(b) {
			return ErrInvalidMaterial
		}
		rank = next
	}
	if rank == 0 {
		return ErrInvalidMaterial
	}
	return nil
}

// The fixed array describes representation properties, not a trusted time.
// No retained copy of the input is made while comparing interval formats.
func checkTimePoint(b []byte) ([7]int, []byte, bool) {
	var shape [7]int
	t := -1
	for i, c := range b {
		if c == 'T' {
			if t >= 0 {
				return shape, nil, false
			}
			t = i
		}
	}
	if t >= 0 {
		kind, year, width, ok := checkTimeDate(b[:t])
		if !ok || kind != 4 && kind != 5 && kind != 7 {
			return shape, nil, false
		}
		shape[0], shape[1], shape[2], shape[3] = 2, kind, year, width
		b = b[t+1:]
	} else if len(b) >= 4 && timeDigits(b[:4]) || len(b) > 0 && (b[0] == '+' || b[0] == '-' || b[len(b)-1] == 'C') {
		kind, year, width, ok := checkTimeDate(b)
		shape[0], shape[1], shape[2], shape[3] = 1, kind, year, width
		return shape, nil, ok
	}
	if len(b) < 2 {
		return shape, nil, false
	}
	zoneAt := len(b)
	for i, c := range b {
		if c == 'Z' || c == '+' || c == '-' {
			zoneAt = i
			break
		}
	}
	zone := b[zoneAt:]
	b = b[:zoneAt]
	if len(zone) > 0 {
		if len(zone) == 1 && zone[0] == 'Z' {
			shape[6] = 1
		} else {
			if (len(zone) != 3 && len(zone) != 6) || (zone[0] != '+' && zone[0] != '-') || !timeDigits(zone[1:3]) {
				return shape, nil, false
			}
			h, m := smallTimeNumber(zone[1:3]), 0
			if len(zone) == 6 {
				if zone[3] != ':' || !timeDigits(zone[4:]) {
					return shape, nil, false
				}
				m = smallTimeNumber(zone[4:])
				if m == 0 || m > 59 {
					return shape, nil, false
				}
			}
			if h > 15 || h == 15 && m > 0 || zone[0] == '-' && h == 0 && m == 0 {
				return shape, nil, false
			}
			shape[6] = 2
		}
	}
	dot := -1
	var fraction []byte
	for i, c := range b {
		if c == '.' {
			if dot >= 0 {
				return shape, nil, false
			}
			dot = i
		}
	}
	if dot >= 0 {
		if !timeDigits(b[dot+1:]) {
			return shape, nil, false
		}
		shape[5] = len(b) - dot - 1
		fraction = b[dot+1:]
		b = b[:dot]
	}
	if len(b) != 2 && len(b) != 5 && len(b) != 8 || !timeDigits(b[:2]) {
		return shape, nil, false
	}
	h, m, s := smallTimeNumber(b[:2]), 0, 0
	shape[4] = 1
	if len(b) >= 5 {
		if b[2] != ':' || !timeDigits(b[3:5]) {
			return shape, nil, false
		}
		m = smallTimeNumber(b[3:5])
		shape[4] = 2
	}
	if len(b) == 8 {
		if b[5] != ':' || !timeDigits(b[6:]) {
			return shape, nil, false
		}
		s = smallTimeNumber(b[6:])
		shape[4] = 3
	}
	if h > 24 || m > 59 || s > 60 || h == 24 && (m != 0 || s != 0) {
		return shape, nil, false
	}
	// End-of-day 24 may retain fractional accuracy, but must still be midnight.
	if h == 24 && dot >= 0 {
		for _, c := range fraction {
			if c != '0' {
				return shape, nil, false
			}
		}
	}
	return shape, zone, true
}
func checkTimeDate(b []byte) (int, int, int, bool) {
	if len(b) == 0 {
		return 0, 0, 0, false
	}
	negative, signed := b[0] == '-', b[0] == '-' || b[0] == '+'
	if signed {
		b = b[1:]
	}
	i, modulo, nonzero := 0, 0, false
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		modulo = (modulo*10 + int(b[i]-'0')) % 400
		nonzero = nonzero || b[i] != '0'
		i++
	}
	century := i < len(b) && b[i] == 'C'
	minimum := 4
	if century {
		minimum = 2
	}
	if !signed && i != minimum || signed && i < minimum || signed && !negative && i == minimum || negative && !nonzero || signed && i > minimum && b[0] == '0' {
		return 0, 0, 0, false
	}
	year, width := 0, 0
	if signed && i > minimum {
		year, width = 3, i
	} else if negative {
		year = 2
	} else if century && smallTimeNumber(b[:i]) < 15 || !century && smallTimeNumber(b[:i]) < 1582 {
		year = 1
	}
	if negative {
		modulo = (400 - modulo) % 400
	}
	leap := modulo%4 == 0 && (modulo%100 != 0 || modulo == 0)
	rest := b[i:]
	if century {
		return 1, year, width, len(rest) == 1
	}
	if len(rest) == 0 {
		return 2, year, width, true
	}
	if rest[0] != '-' {
		return 0, 0, 0, false
	}
	rest = rest[1:]
	if len(rest) == 2 && timeDigits(rest) {
		m := smallTimeNumber(rest)
		return 3, year, width, m >= 1 && m <= 12
	}
	if len(rest) == 5 && rest[2] == '-' && timeDigits(rest[:2]) && timeDigits(rest[3:]) {
		m, d := smallTimeNumber(rest[:2]), smallTimeNumber(rest[3:])
		if m < 1 || m > 12 {
			return 0, 0, 0, false
		}
		days := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
		if leap {
			days[1] = 29
		}
		return 4, year, width, d >= 1 && d <= days[m-1]
	}
	if len(rest) == 3 && timeDigits(rest) {
		d, maximum := smallTimeNumber(rest), 365
		if leap {
			maximum++
		}
		return 5, year, width, d >= 1 && d <= maximum
	}
	if (len(rest) == 3 || len(rest) == 5) && rest[0] == 'W' && timeDigits(rest[1:3]) {
		w := smallTimeNumber(rest[1:3])
		if w < 1 || w > 53 {
			return 0, 0, 0, false
		}
		if w == 53 {
			y := modulo
			if y == 0 {
				y = 400
			}
			p := y - 1
			weekday := (365*p + p/4 - p/100 + p/400 + 1) % 7
			if weekday != 4 && (weekday != 3 || !leap) {
				return 0, 0, 0, false
			}
		}
		if len(rest) == 3 {
			return 6, year, width, true
		}
		return 7, year, width, rest[3] == '-' && rest[4] >= '1' && rest[4] <= '7'
	}
	return 0, 0, 0, false
}
func timeDigits(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Callers limit this conversion to the small fixed-width calendar components.
func smallTimeNumber(b []byte) int {
	n := 0
	for _, c := range b {
		n = n*10 + int(c-'0')
	}
	return n
}

// Validate DER REAL representation without converting its potentially unbounded
// mantissa or exponent, and without adding a floating-point value policy.
func checkGeneralReal(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if b[0]&0x80 != 0 {
		if b[0]&0x3c != 0 {
			return ErrInvalidMaterial
		}
		n, at := int(b[0]&3)+1, 1
		if n == 4 {
			if len(b) < 2 || b[1] <= 3 {
				return ErrInvalidMaterial
			}
			n, at = int(b[1]), 2
		}
		if len(b) <= at+n || !minimalInteger(b[at:at+n]) || b[at+n] == 0 || b[len(b)-1]&1 == 0 {
			return ErrInvalidMaterial
		}
		return nil
	}
	if b[0] >= 0x40 && b[0] <= 0x43 && len(b) == 1 {
		return nil
	}
	if b[0] != 3 {
		return ErrInvalidMaterial
	}
	s := b[1:]
	if len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	if len(s) == 0 || s[0] < '1' || s[0] > '9' {
		return ErrInvalidMaterial
	}
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i+2 >= len(s) || s[i-1] == '0' || s[i] != '.' || s[i+1] != 'E' {
		return ErrInvalidMaterial
	}
	exponent := s[i+2:]
	if bytes.Equal(exponent, []byte("+0")) {
		return nil
	}
	if len(exponent) > 0 && exponent[0] == '-' {
		exponent = exponent[1:]
	}
	if len(exponent) == 0 || exponent[0] < '1' || exponent[0] > '9' {
		return ErrInvalidMaterial
	}
	for _, c := range exponent {
		if c < '0' || c > '9' {
			return ErrInvalidMaterial
		}
	}
	return nil
}

// EXTERNAL has optional references/descriptor and one required encoding choice.
// Payload transfer syntax and legacy ObjectDescriptor repertoire are not resolved.
func checkExternalValue(body []byte) error {
	if len(body) > 0 && body[0] == 6 {
		v, e := takeElement(&body, 6)
		if e != nil || !minimalOIDContent(v.content) {
			return ErrInvalidMaterial
		}
	}
	if len(body) > 0 && body[0] == 2 {
		v, e := takeElement(&body, 2)
		if e != nil || !minimalInteger(v.content) {
			return ErrInvalidMaterial
		}
	}
	if len(body) > 0 && body[0] == 7 {
		if _, e := takeElement(&body, 7); e != nil {
			return e
		}
	}
	v, tail, e := readElement(body)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	switch v.tag {
	case 0xa0:
		_, tail, e := readElement(v.content)
		if e != nil || len(tail) != 0 {
			return ErrInvalidMaterial
		}
	case 0x81:
	case 0x82:
		if !bitStringShape(v.content) {
			return ErrInvalidMaterial
		}
	default:
		return ErrInvalidMaterial
	}
	// Nested single-ASN1-type values are checked by the iterative outer walk.
	return nil
}

// EMBEDDED PDV and CHARACTER STRING use the same automatically tagged
// identification plus data-octets container. No vendor transfer syntax is decoded.
func checkEmbeddedValue(body []byte) error {
	explicit, e := takeElement(&body, 0xa0)
	if e != nil {
		return e
	}
	id, tail, e := readElement(explicit.content)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	fields := id.content
	switch id.tag {
	case 0xa0:
		for _, tag := range []byte{0x80, 0x81} {
			v, e := takeElement(&fields, tag)
			if e != nil || !minimalOIDContent(v.content) {
				return ErrInvalidMaterial
			}
		}
		if len(fields) != 0 {
			return ErrInvalidMaterial
		}
	case 0x81, 0x84:
		if !minimalOIDContent(fields) {
			return ErrInvalidMaterial
		}
	case 0x82:
		if !minimalInteger(fields) {
			return ErrInvalidMaterial
		}
	case 0xa3:
		v, e := takeElement(&fields, 0x80)
		if e != nil || !minimalInteger(v.content) {
			return ErrInvalidMaterial
		}
		v, e = takeElement(&fields, 0x81)
		if e != nil || !minimalOIDContent(v.content) || len(fields) != 0 {
			return ErrInvalidMaterial
		}
	case 0x85:
		if len(fields) != 0 {
			return ErrInvalidMaterial
		}
	default:
		return ErrInvalidMaterial
	}
	if _, e = takeElement(&body, 0x82); e != nil || len(body) != 0 {
		return ErrInvalidMaterial
	}
	return nil
}
func checkUniversalString(v derElement) error {
	if v.tag == 26 {
		for _, c := range v.content {
			if c < 32 || c > 126 {
				return ErrInvalidMaterial
			}
		}
		return nil
	}
	if v.tag == 28 {
		if len(v.content)%4 != 0 {
			return ErrInvalidMaterial
		}
		for i := 0; i < len(v.content); i += 4 {
			n := uint32(v.content[i])<<24 | uint32(v.content[i+1])<<16 | uint32(v.content[i+2])<<8 | uint32(v.content[i+3])
			if n > 0x10ffff || (n >= 0xd800 && n <= 0xdfff) {
				return ErrInvalidMaterial
			}
		}
		return nil
	}
	var decoded string
	tail, e := asn1.Unmarshal(v.full, &decoded)
	if e != nil || len(tail) != 0 {
		return ErrInvalidMaterial
	}
	return nil
}
func checkImplicitString(tag byte, b []byte) error {
	for _, v := range b {
		switch tag {
		case 18:
			if v != ' ' && (v < '0' || v > '9') {
				return ErrInvalidMaterial
			}
		case 19:
			// Same PrintableString interpretation as the pinned Go parsers, including
			// their legacy asterisk/ampersand allowances; not a naming policy.
			if !((v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z') || (v >= '0' && v <= '9') ||
				(v >= 39 && v <= 41) || (v >= 43 && v <= 47) || v == ' ' || v == ':' || v == '=' || v == '?' || v == '*' || v == '&') {
				return ErrInvalidMaterial
			}
		case 20: // Pinned Go interprets TeletexString as an 8-bit string.
		case 22:
			if v >= 128 {
				return ErrInvalidMaterial
			}
		default:
			return ErrInvalidMaterial
		}
	}
	return nil
}

// RFC5280 Appendix A.1 container grammars only. All such issuer-bearing AKIs
// are still rejected; length/range, matching and authorization policies are not
// inferred from these syntax checks.
func checkORAddress(body []byte) error {
	standard, e := takeElement(&body, 0x30)
	if e != nil {
		return e
	}
	if e = checkX400Standard(standard.content); e != nil {
		return e
	}
	if len(body) > 0 && body[0] == 0x30 {
		domain, e := takeElement(&body, 0x30)
		if e != nil {
			return e
		}
		if e = checkX400Domain(domain.content, 19); e != nil {
			return e
		}
	}
	if len(body) > 0 && body[0] == 0x31 {
		ext, e := takeElement(&body, 0x31)
		if e != nil {
			return e
		}
		if e = checkX400Extensions(ext.content); e != nil {
			return e
		}
	}
	if len(body) != 0 {
		return ErrInvalidMaterial
	}
	return nil
}
func checkX400Standard(body []byte) error {
	previous := -1
	for len(body) > 0 {
		v, tail, e := readElement(body)
		if e != nil {
			return e
		}
		index := -1
		for i, tag := range []byte{0x61, 0x62, 0x80, 0x81, 0xa2, 0x83, 0x84, 0xa5, 0xa6} {
			if v.tag == tag {
				index = i
				break
			}
		}
		if index <= previous {
			return ErrInvalidMaterial
		}
		previous = index
		switch v.tag {
		case 0x61, 0x62, 0xa2:
			if e = checkNumericPrintableChoice(v.content); e != nil {
				return e
			}
		case 0x80, 0x84:
			if e = checkImplicitString(18, v.content); e != nil {
				return e
			}
		case 0x81, 0x83:
			if e = checkImplicitString(19, v.content); e != nil {
				return e
			}
		case 0xa5:
			if e = checkX400Personal(v.content, 19); e != nil {
				return e
			}
		case 0xa6:
			if e = checkX400StringSequence(v.content, 19); e != nil {
				return e
			}
		}
		body = tail
	}
	return nil
}
func checkNumericPrintableChoice(raw []byte) error {
	v, tail, e := readElement(raw)
	if e != nil || len(tail) != 0 || (v.tag != 18 && v.tag != 19) {
		return ErrInvalidMaterial
	}
	return checkGeneralValue(v)
}
func checkX400Personal(body []byte, tag byte) error {
	surname, e := takeElement(&body, 0x80)
	if e != nil {
		return e
	}
	if e = checkImplicitString(tag, surname.content); e != nil {
		return e
	}
	previous := byte(0x80)
	for len(body) > 0 {
		v, tail, e := readElement(body)
		if e != nil {
			return e
		}
		if v.tag <= previous || v.tag > 0x83 {
			return ErrInvalidMaterial
		}
		previous = v.tag
		if e = checkImplicitString(tag, v.content); e != nil {
			return e
		}
		body = tail
	}
	return nil
}
func checkX400StringSequence(body []byte, tag byte) error {
	for len(body) > 0 {
		v, e := takeElement(&body, tag)
		if e != nil {
			return e
		}
		if e = checkGeneralValue(v); e != nil {
			return e
		}
	}
	return nil
}
func checkX400Domain(body []byte, tag byte) error {
	for len(body) > 0 {
		v, e := takeElement(&body, 0x30)
		if e != nil {
			return e
		}
		fields := v.content
		for i := 0; i < 2; i++ {
			s, e := takeElement(&fields, tag)
			if e != nil {
				return e
			}
			if e = checkGeneralValue(s); e != nil {
				return e
			}
		}
		if len(fields) != 0 {
			return ErrInvalidMaterial
		}
	}
	return nil
}
func checkX400Extensions(body []byte) error {
	var previous []byte
	for len(body) > 0 {
		v, e := takeElement(&body, 0x30)
		if e != nil {
			return e
		}
		if previous != nil && derLess(v.full, previous) {
			return ErrInvalidMaterial
		}
		previous = v.full
		fields := v.content
		kind, e := takeElement(&fields, 0x80)
		if e != nil || !minimalInteger(kind.content) {
			return ErrInvalidMaterial
		}
		explicit, e := takeElement(&fields, 0xa1)
		if e != nil || len(fields) != 0 {
			return ErrInvalidMaterial
		}
		value, tail, e := readElement(explicit.content)
		if e != nil || len(tail) != 0 {
			return ErrInvalidMaterial
		}
		if e = checkGeneralValue(value); e != nil {
			return e
		}
		if len(kind.content) == 1 {
			if e = checkX400Attribute(kind.content[0], value); e != nil {
				return e
			}
		}
	}
	return nil
}
func checkX400Attribute(kind byte, v derElement) error {
	switch kind {
	case 1, 7:
		if v.tag != 19 {
			return ErrInvalidMaterial
		}
		return checkGeneralValue(v)
	case 2, 3:
		if v.tag != 20 {
			return ErrInvalidMaterial
		}
		return checkGeneralValue(v)
	case 4:
		if v.tag != 0x31 {
			return ErrInvalidMaterial
		}
		return checkX400Personal(v.content, 20)
	case 5:
		if v.tag != 0x30 {
			return ErrInvalidMaterial
		}
		return checkX400StringSequence(v.content, 20)
	case 6:
		if v.tag != 0x30 {
			return ErrInvalidMaterial
		}
		return checkX400Domain(v.content, 20)
	case 8, 9:
		return checkNumericPrintableChoice(v.full)
	case 10, 11, 12, 13, 14, 15, 17, 18, 19, 20, 21:
		if v.tag != 0x31 {
			return ErrInvalidMaterial
		}
		return checkX400Postal(v.content, false)
	case 16:
		if v.tag != 0x31 {
			return ErrInvalidMaterial
		}
		return checkX400Postal(v.content, true)
	case 22:
		if v.tag == 0xa0 {
			return checkX400Presentation(v.content)
		}
		if v.tag != 0x30 {
			return ErrInvalidMaterial
		}
		body := v.content
		number, e := takeElement(&body, 0x80)
		if e != nil {
			return e
		}
		if e = checkImplicitString(18, number.content); e != nil {
			return e
		}
		if len(body) > 0 && body[0] == 0x81 {
			sub, e := takeElement(&body, 0x81)
			if e != nil {
				return e
			}
			if e = checkImplicitString(18, sub.content); e != nil {
				return e
			}
		}
		if len(body) != 0 {
			return ErrInvalidMaterial
		}
		return nil
	case 23:
		if v.tag != 2 || !minimalInteger(v.content) {
			return ErrInvalidMaterial
		}
	}
	return nil // Unknown attribute types retain consumed/minimal ANY, not schema claims.
}
func checkX400Postal(body []byte, address bool) error {
	previous := -1
	for len(body) > 0 {
		v, tail, e := readElement(body)
		if e != nil {
			return e
		}
		index := -1
		if v.tag == 20 {
			index = 1
		}
		if (!address && v.tag == 19) || (address && v.tag == 0x30) {
			index = 0
		}
		if index <= previous {
			return ErrInvalidMaterial
		}
		previous = index
		if address && index == 0 {
			if e = checkX400StringSequence(v.content, 19); e != nil {
				return e
			}
		} else if e = checkGeneralValue(v); e != nil {
			return e
		}
		body = tail
	}
	return nil
}
func checkX400Presentation(body []byte) error {
	for i := byte(0); i < 3; i++ {
		if len(body) > 0 && body[0] == 0xa0+i {
			selector, e := takeElement(&body, 0xa0+i)
			if e != nil {
				return e
			}
			if _, e = singleElement(selector.content, 4); e != nil {
				return e
			}
		}
	}
	explicit, e := takeElement(&body, 0xa3)
	if e != nil || len(body) != 0 {
		return ErrInvalidMaterial
	}
	addresses, e := singleElement(explicit.content, 0x31)
	if e != nil {
		return e
	}
	var previous []byte
	for rest := addresses.content; len(rest) > 0; {
		v, e := takeElement(&rest, 4)
		if e != nil {
			return e
		}
		if previous != nil && derLess(v.full, previous) {
			return ErrInvalidMaterial
		}
		previous = v.full
	}
	return nil
}
func checkBasicConstraints(b []byte) (bool, uint8, error) {
	seq, e := singleElement(b, 0x30)
	if e != nil {
		return false, 0, e
	}
	body := seq.content
	cA := false
	if len(body) != 0 && body[0] == 1 {
		flag, err := takeElement(&body, 1)
		if err != nil || len(flag.content) != 1 || flag.content[0] != 255 {
			return false, 0, ErrInvalidMaterial
		}
		cA = true
	}
	hasPath := len(body) != 0
	var n derElement
	if hasPath {
		n, e = takeElement(&body, 2)
		if e != nil || len(body) != 0 || !minimalInteger(n.content) {
			return false, 0, ErrInvalidMaterial
		}
	}
	// Complete structural consumption precedes an interpretable profile rejection.
	if !cA {
		return false, 0, ErrUnsupportedCAProfile
	}
	if !hasPath {
		return false, 0, nil
	}
	if n.content[0]&128 != 0 || len(n.content) > 2 {
		return false, 0, ErrUnsupportedCAProfile
	}
	value := 0
	for _, x := range n.content {
		value = value*256 + int(x)
	}
	if value > 255 {
		return false, 0, ErrUnsupportedCAProfile
	}
	return true, uint8(value), nil
}
func checkKeyUsage(b []byte) (bool, error) {
	bit, e := singleElement(b, 3)
	if e != nil {
		return false, e
	}
	if !bitStringShape(bit.content) {
		return false, ErrInvalidMaterial
	}
	if len(bit.content) > 1 {
		last := bit.content[len(bit.content)-1]
		if last == 0 {
			return false, ErrInvalidMaterial
		}
		unused := byte(0)
		for last&1 == 0 {
			unused++
			last >>= 1
		}
		if unused != bit.content[0] {
			return false, ErrInvalidMaterial
		}
	}
	if bytes.Equal(b, []byte{3, 2, 2, 4}) {
		return false, nil
	}
	if bytes.Equal(b, []byte{3, 2, 1, 6}) {
		return true, nil
	}
	return false, ErrUnsupportedCAProfile
}
