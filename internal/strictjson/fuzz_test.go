package strictjson_test

import (
	"bytes"
	"errors"
	"testing"

	"talenro.local/platform/internal/strictjson"
)

func FuzzDecode(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{"name":"ok"}`),
		[]byte(`{"unknown":"SECRET-CANARY"}`),
		[]byte(`{"name":"a","name":"b"}`),
		[]byte(`{"nested":{"enabled":true,"enabled":false}}`),
		{'{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 0xff, '"', '}'},
		[]byte(`{"name":"a"}{"name":"b"}`),
		[]byte(`[[[[[[[[[[[[[[[[[0]]]]]]]]]]]]]]]]]`),
	} {
		f.Add(seed)
	}

	finiteErrors := []error{
		strictjson.ErrInvalidArgument,
		strictjson.ErrTooLarge,
		strictjson.ErrInvalidUTF8,
		strictjson.ErrInvalidJSON,
		strictjson.ErrDuplicateMember,
		strictjson.ErrTooDeep,
		strictjson.ErrUnknownMember,
		strictjson.ErrTrailingData,
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		var target request
		err := strictjson.Decode(bytes.NewReader(body), maxBodyBytes, &target)
		if err == nil {
			withTrailing := make([]byte, 0, len(body)+4)
			withTrailing = append(withTrailing, body...)
			withTrailing = append(withTrailing, []byte(`null`)...)
			if trailingErr := strictjson.Decode(bytes.NewReader(withTrailing), maxBodyBytes, &target); trailingErr == nil {
				t.Fatal("Decode() accepted a second JSON value")
			}
			return
		}

		known := false
		for _, sentinel := range finiteErrors {
			known = known || errors.Is(err, sentinel)
		}
		if !known {
			t.Fatalf("Decode() returned non-finite error %q", err)
		}
		if len(body) >= 8 && bytes.Contains([]byte(err.Error()), body) {
			t.Fatalf("Decode() error disclosed input bytes")
		}
		if bytes.Contains([]byte(err.Error()), []byte("SECRET-CANARY")) {
			t.Fatal("Decode() error disclosed canary")
		}
	})
}
