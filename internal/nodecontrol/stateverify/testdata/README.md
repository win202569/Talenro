Test-only fixed vectors. Go standard Ed25519, seeds: 32 bytes all zero except byte 0 = 1 (root), 2 (desired). Never production keys. ASCII literal payloads with sorted JSON keys; ordinary signatures over complete domain + NUL + payload. Transcript hex is decoded before comparison. Generated independently of stateformat/stateverify.

