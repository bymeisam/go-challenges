package main

import (
	"errors"

	"github.com/google/uuid"
)

func GenerateUUIDv4() (string, error) {
	id := uuid.New()
	return id.String(), nil
}

func ParseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

func GenerateUUIDFromName(namespace uuid.UUID, name string) string {
	id := uuid.NewSHA1(namespace, []byte(name))
	return id.String()
}

func GetUUIDVersion(uuidStr string) (int, error) {
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return 0, err
	}
	
	version := id.Version()
	if version == 0 {
		return 0, errors.New("invalid UUID version")
	}
	
	return int(version), nil
}

func UUIDToBytes(uuidStr string) ([]byte, error) {
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

func UUIDFromBytes(b []byte) (string, error) {
	id, err := uuid.FromBytes(b)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func main() {}
