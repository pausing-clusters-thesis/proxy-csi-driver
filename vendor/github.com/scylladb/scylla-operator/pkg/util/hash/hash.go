// Copyright (C) 2021 ScyllaDB

package hash

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
)

func HashObjects(objs ...interface{}) (string, error) {
	hasher := sha512.New()
	encoder := json.NewEncoder(hasher)
	for _, obj := range objs {
		if err := encoder.Encode(obj); err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

func HashBytes(buf []byte) (string, error) {
	hasher := sha512.New()

	_, err := hasher.Write(buf)
	if err != nil {
		return "", fmt.Errorf("can't write bytes to hasher: %w", err)
	}
	return string(hasher.Sum(nil)), nil
}

func HashObjectFNV64a(objs ...interface{}) (uint64, error) {
	hasher := fnv.New64a()
	encoder := json.NewEncoder(hasher)
	for _, obj := range objs {
		if err := encoder.Encode(obj); err != nil {
			return 0, err
		}
	}

	return hasher.Sum64(), nil
}
