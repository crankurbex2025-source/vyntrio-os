// Package setting defines system settings domain types and validation.
package setting

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const NamespaceSystem = "system"

const (
	KeyTimezone = "timezone"
	KeyHostname = "hostname"
)

// ValueType is the persisted scalar type for a setting value.
type ValueType string

const (
	ValueTypeString ValueType = "string"
	ValueTypeInt    ValueType = "int"
	ValueTypeBool   ValueType = "bool"
)

// Setting is a namespaced configuration value.
type Setting struct {
	Namespace string
	Key       string
	Value     string
	ValueType ValueType
	UpdatedAt string
}

var (
	ErrInvalidNamespace = errors.New("invalid settings namespace")
	ErrInvalidKey       = errors.New("invalid settings key")
	ErrInvalidValueType = errors.New("invalid settings value type")
	ErrInvalidValue     = errors.New("invalid settings value")
)

var allowedKeys = map[string]struct{}{
	KeyTimezone: {},
	KeyHostname: {},
}

// ValidateNamespace ensures only the system namespace is used in Slice 2b.1.
func ValidateNamespace(namespace string) error {
	if namespace != NamespaceSystem {
		return fmt.Errorf("%w: %q", ErrInvalidNamespace, namespace)
	}
	return nil
}

// ValidateKey ensures the key is one of the allowed system keys.
func ValidateKey(namespace, key string) error {
	if err := ValidateNamespace(namespace); err != nil {
		return err
	}
	if _, ok := allowedKeys[key]; !ok {
		return fmt.Errorf("%w: %q", ErrInvalidKey, key)
	}
	return nil
}

// ValidateValueType ensures the value type is supported.
func ValidateValueType(valueType ValueType) error {
	switch valueType {
	case ValueTypeString, ValueTypeInt, ValueTypeBool:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidValueType, valueType)
	}
}

// ValidateValue validates a value for a known system key.
func ValidateValue(key, value string) error {
	switch key {
	case KeyTimezone:
		return validateTimezone(value)
	case KeyHostname:
		return validateHostname(value)
	default:
		return fmt.Errorf("%w: %q", ErrInvalidKey, key)
	}
}

// Validate checks namespace, key, value type, and value for a setting.
func (s Setting) Validate() error {
	if err := ValidateKey(s.Namespace, s.Key); err != nil {
		return err
	}
	if err := ValidateValueType(s.ValueType); err != nil {
		return err
	}
	if s.ValueType != ValueTypeString {
		return fmt.Errorf("%w: system settings require string values in Slice 2b.1", ErrInvalidValueType)
	}
	return ValidateValue(s.Key, s.Value)
}

func validateTimezone(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%w: timezone must not be empty", ErrInvalidValue)
	}
	if utf8.RuneCountInString(trimmed) > 64 {
		return fmt.Errorf("%w: timezone exceeds maximum length", ErrInvalidValue)
	}
	return nil
}

func validateHostname(value string) error {
	_, err := ValidateInstanceDisplayName(value)
	return err
}
