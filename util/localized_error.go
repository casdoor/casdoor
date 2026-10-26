// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"fmt"
	"strings"
)

// MessageError retains a first-party template without depending on i18n or conf.
// Low-level validation can therefore keep provenance without an import cycle.
type MessageError struct {
	key       string
	args      []interface{}
	formatted bool
	english   error
	cause     error
}

func messageTemplate(key string) string {
	_, template, found := strings.Cut(key, ":")
	if !found {
		return key
	}
	return template
}

// MessageErrorf has fmt.Errorf's English formatting and %w cause semantics.
func MessageErrorf(key string, args ...interface{}) error {
	english := fmt.Errorf(messageTemplate(key), args...)
	return &MessageError{key: key, args: append([]interface{}(nil), args...), formatted: true, english: english, cause: english}
}

// NewMessageError, like errors.New, does not interpret percent signs.
func NewMessageError(key string) error {
	english := errors.New(messageTemplate(key))
	return &MessageError{key: key, english: english, cause: english}
}

func WrapMessageError(cause error, key string, args ...interface{}) error {
	value := MessageErrorf(key, args...).(*MessageError)
	if cause != nil {
		value.cause = errors.Join(value.english, cause)
	}
	return value
}

func (e *MessageError) Error() string { return e.english.Error() }

func (e *MessageError) Unwrap() error { return e.cause }

type joinedMessageErrors struct {
	separator string
	errors    []error
}

func JoinMessageErrors(separator string, values ...error) error {
	items := make([]error, 0, len(values))
	for _, value := range values {
		if value != nil {
			items = append(items, value)
		}
	}
	if len(items) == 0 {
		return nil
	}
	return &joinedMessageErrors{separator: separator, errors: items}
}

func (e *joinedMessageErrors) Error() string {
	parts := make([]string, len(e.errors))
	for i, value := range e.errors {
		parts[i] = value.Error()
	}
	return strings.Join(parts, e.separator)
}

func (e *joinedMessageErrors) Unwrap() []error {
	return append([]error(nil), e.errors...)
}

func IsMessageError(err error) bool {
	switch value := err.(type) {
	case *MessageError:
		return value != nil
	case *joinedMessageErrors:
		for _, item := range value.errors {
			if IsMessageError(item) {
				return true
			}
		}
	}
	return false
}

// FormatMessageError only translates this package's own message types. It never
// unwraps an unknown error or interprets an argument as a translation key.
func FormatMessageError(err error, translate func(string) string) string {
	if err == nil {
		return ""
	}
	switch value := err.(type) {
	case *MessageError:
		if !value.formatted {
			return translate(value.key)
		}
		args := append([]interface{}(nil), value.args...)
		for i, argument := range args {
			if nested, ok := argument.(error); ok && IsMessageError(nested) {
				args[i] = errors.New(FormatMessageError(nested, translate))
			}
		}
		return fmt.Errorf(translate(value.key), args...).Error()
	case *joinedMessageErrors:
		parts := make([]string, len(value.errors))
		for i, item := range value.errors {
			parts[i] = FormatMessageError(item, translate)
		}
		return strings.Join(parts, value.separator)
	default:
		return err.Error()
	}
}
