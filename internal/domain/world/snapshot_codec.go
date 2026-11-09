package world

import (
	"bytes"
	"encoding/gob"
)

// EncodeSnapshot — gob-кодирование снапшота для сети.
func EncodeSnapshot(s *Snapshot) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeSnapshot — обратное преобразование.
func DecodeSnapshot(data []byte) (Snapshot, error) {
	var s Snapshot
	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&s)
	return s, err
}
