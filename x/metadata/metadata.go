package metadata

import (
	"strings"

	"github.com/168yy/netx/core/metadata"
)

type mapMetadata map[string]any

func NewMetadata(m map[string]any) metadata.IMetaData {
	if len(m) == 0 {
		return nil
	}
	md := make(map[string]any)
	for k, v := range m {
		md[strings.ToLower(k)] = v
	}
	return mapMetadata(md)
}

func (m mapMetadata) IsExists(key string) bool {
	_, ok := m[strings.ToLower(key)]
	return ok
}

func (m mapMetadata) Set(key string, value any) {
	m[strings.ToLower(key)] = value
}

func (m mapMetadata) Get(key string) any {
	if m != nil {
		return m[strings.ToLower(key)]
	}
	return nil
}
