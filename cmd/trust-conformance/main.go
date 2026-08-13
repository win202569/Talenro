// Command trust-conformance verifies and durably activates one file-based
// Talenro trust-envelope fixture without printing secret material.
package main

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"talenro.local/platform/internal/secret"
	"talenro.local/platform/internal/trustclient"
)

const (
	maximumCLIEnvelopeBytes = 1 << 20
	maximumCLIMetadataBytes = 64 << 10
	maximumCLIPrivateBytes  = 128
	maximumCLIJSONDepth     = 16
)

type conformanceMetadataFile struct {
	SignedMetadata        json.RawMessage `json:"signed_metadata"`
	TrustedRoots          []rootFile      `json:"trusted_roots"`
	HighestTrusted        string          `json:"highest_trusted_version"`
	ExpectedAudience      string          `json:"expected_audience"`
	ExpectedBundleLocator string          `json:"expected_bundle_locator"`
}

type rootFile struct {
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(arguments []string, output io.Writer) int {
	flags := flag.NewFlagSet("trust-conformance", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	envelopePath := flags.String("envelope", "", "envelope file")
	metadataPath := flags.String("metadata", "", "metadata file")
	privateKeyPath := flags.String("private-key", "", "recipient private-key file")
	stateDirectory := flags.String("state-dir", "", "durable state directory")
	if flags.Parse(arguments) != nil || flags.NArg() != 0 || *envelopePath == "" || *metadataPath == "" || *privateKeyPath == "" || *stateDirectory == "" {
		return emit(output, "arguments")
	}
	envelope, err := readBoundedFile(*envelopePath, maximumCLIEnvelopeBytes)
	if err != nil {
		return emit(output, "envelope")
	}
	defer clear(envelope)
	metadataBody, err := readBoundedFile(*metadataPath, maximumCLIMetadataBytes)
	if err != nil {
		return emit(output, "metadata")
	}
	defer clear(metadataBody)
	privateBody, err := readBoundedFile(*privateKeyPath, maximumCLIPrivateBytes)
	if err != nil {
		return emit(output, "private_key")
	}
	privateSecret := secret.NewBytes(privateBody)
	clear(privateBody)
	privateOwner := privateSecret.Take()
	defer privateOwner.Clear()
	encodedPrivate := privateOwner.Copy()
	defer clear(encodedPrivate)
	trimmedPrivate := bytes.TrimSpace(encodedPrivate)
	if len(trimmedPrivate) == 0 || bytes.ContainsAny(trimmedPrivate, " \t\r\n") {
		return emit(output, "private_key")
	}
	privateBytes, err := base64.RawURLEncoding.DecodeString(string(trimmedPrivate))
	if err != nil || len(privateBytes) != 32 || base64.RawURLEncoding.EncodeToString(privateBytes) != string(trimmedPrivate) {
		clear(privateBytes)
		return emit(output, "private_key")
	}
	privateKey, err := ecdh.X25519().NewPrivateKey(privateBytes)
	clear(privateBytes)
	if err != nil {
		return emit(output, "private_key")
	}

	metadata, expected, err := parseMetadataFile(metadataBody)
	if err != nil {
		return emit(output, "metadata")
	}
	store, err := trustclient.NewDirectoryStore(*stateDirectory)
	if err != nil {
		return emit(output, "state")
	}
	_, err = trustclient.VerifyAndStage(
		context.Background(), envelope, privateKey, metadata, expected, store, time.Now().UTC(), 120*time.Second,
	)
	if err != nil {
		return emit(output, category(err))
	}
	_, _ = io.WriteString(output, "verified\n")
	return 0
}

func parseMetadataFile(body []byte) (trustclient.TrustMetadata, trustclient.Expected, error) {
	if err := validateFiniteJSON(body, maximumCLIMetadataBytes); err != nil {
		return trustclient.TrustMetadata{}, trustclient.Expected{}, err
	}
	var value conformanceMetadataFile
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return trustclient.TrustMetadata{}, trustclient.Expected{}, err
	}
	if value.HighestTrusted == "" || len(value.HighestTrusted) > 20 || len(value.HighestTrusted) > 1 && value.HighestTrusted[0] == '0' ||
		len(value.TrustedRoots) == 0 || len(value.TrustedRoots) > 64 {
		return trustclient.TrustMetadata{}, trustclient.Expected{}, trustclient.ErrInvalidArgument
	}
	highest, err := strconv.ParseUint(value.HighestTrusted, 10, 64)
	if err != nil || highest > math.MaxInt64 {
		return trustclient.TrustMetadata{}, trustclient.Expected{}, trustclient.ErrInvalidArgument
	}
	roots := make(map[string]ed25519.PublicKey, len(value.TrustedRoots))
	for _, root := range value.TrustedRoots {
		publicKey, err := base64.RawURLEncoding.DecodeString(root.PublicKey)
		if err != nil || len(publicKey) != ed25519.PublicKeySize || base64.RawURLEncoding.EncodeToString(publicKey) != root.PublicKey {
			clear(publicKey)
			return trustclient.TrustMetadata{}, trustclient.Expected{}, trustclient.ErrInvalidArgument
		}
		if _, duplicate := roots[root.KeyID]; duplicate {
			clear(publicKey)
			return trustclient.TrustMetadata{}, trustclient.Expected{}, trustclient.ErrInvalidArgument
		}
		roots[root.KeyID] = ed25519.PublicKey(publicKey)
	}
	metadata, err := trustclient.NewTrustMetadata(value.SignedMetadata, roots, highest)
	for keyID := range roots {
		clear(roots[keyID])
	}
	if err != nil {
		return trustclient.TrustMetadata{}, trustclient.Expected{}, err
	}
	locator, err := base64.RawURLEncoding.DecodeString(value.ExpectedBundleLocator)
	if err != nil || len(locator) != 32 || base64.RawURLEncoding.EncodeToString(locator) != value.ExpectedBundleLocator {
		clear(locator)
		return trustclient.TrustMetadata{}, trustclient.Expected{}, trustclient.ErrInvalidArgument
	}
	var locatorArray [32]byte
	copy(locatorArray[:], locator)
	clear(locator)
	expected, err := trustclient.NewExpected(value.ExpectedAudience, locatorArray)
	if err != nil {
		return trustclient.TrustMetadata{}, trustclient.Expected{}, err
	}
	return metadata, expected, nil
}

func readBoundedFile(path string, maximum int64) ([]byte, error) {
	if len(path) == 0 || len(path) > 4096 || !utf8.ValidString(path) {
		return nil, trustclient.ErrInvalidArgument
	}
	file, err := os.Open(path) //nolint:gosec // The CLI intentionally opens the explicit user-selected fixture path.
	if err != nil {
		return nil, trustclient.ErrInvalidArgument
	}
	defer func() { _ = file.Close() }()
	information, err := file.Stat()
	if err != nil || !information.Mode().IsRegular() || information.Size() < 1 || information.Size() > maximum {
		return nil, trustclient.ErrInvalidArgument
	}
	body, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || len(body) < 1 || int64(len(body)) > maximum {
		clear(body)
		return nil, trustclient.ErrInvalidArgument
	}
	return body, nil
}

func category(err error) string {
	switch {
	case errors.Is(err, trustclient.ErrEnvelope):
		return "envelope"
	case errors.Is(err, trustclient.ErrMetadata), errors.Is(err, trustclient.ErrMetadataRollback):
		return "metadata"
	case errors.Is(err, trustclient.ErrRecipient):
		return "recipient"
	case errors.Is(err, trustclient.ErrSignature):
		return "signature"
	case errors.Is(err, trustclient.ErrPayload):
		return "payload"
	case errors.Is(err, trustclient.ErrRollback):
		return "rollback"
	case errors.Is(err, trustclient.ErrStore), errors.Is(err, trustclient.ErrStateNotFound):
		return "state"
	default:
		return "arguments"
	}
}

func emit(output io.Writer, value string) int {
	_, _ = io.WriteString(output, value+"\n")
	return 2
}

func validateFiniteJSON(body []byte, maximum int) error {
	if len(body) < 2 || len(body) > maximum || !utf8.Valid(body) {
		return trustclient.ErrInvalidArgument
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := scanJSON(decoder, 0); err != nil {
		return trustclient.ErrInvalidArgument
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return trustclient.ErrInvalidArgument
	}
	return nil
}

func scanJSON(decoder *json.Decoder, depth int) error {
	if depth > maximumCLIJSONDepth {
		return trustclient.ErrInvalidArgument
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			nameToken, err := decoder.Token()
			name, ok := nameToken.(string)
			if err != nil || !ok {
				return trustclient.ErrInvalidArgument
			}
			if _, duplicate := seen[name]; duplicate {
				return trustclient.ErrInvalidArgument
			}
			seen[name] = struct{}{}
			if err := scanJSON(decoder, depth+1); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return trustclient.ErrInvalidArgument
		}
	case '[':
		for decoder.More() {
			if err := scanJSON(decoder, depth+1); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return trustclient.ErrInvalidArgument
		}
	default:
		return trustclient.ErrInvalidArgument
	}
	return nil
}
