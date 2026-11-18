// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
)

var (
	formFillControlKeys   = map[string]bool{"id": true, "name": true, "class": true, "type": true, "index": true}
	formFillFillMethods   = map[string]bool{"": true, "Keyboard": true}
	formFillKeyTokenRegex = regexp.MustCompile(`\{([^{}]*)\}`)
	formFillKeyTokens     = map[string]bool{"username": true, "password": true, "TAB": true, "ENTER": true, "WAIT": true}
)

func normalizeProgramPath(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), `\`, "/"))
}

// matchProgramPath tells whether an executable is the program: a pattern without a directory is the file name, a
// pattern with one is the whole path, where * matches anything
func matchProgramPath(pattern string, program string) bool {
	pattern = normalizeProgramPath(pattern)
	program = normalizeProgramPath(program)
	if pattern == "" || program == "" {
		return false
	}
	if !strings.Contains(pattern, "/") {
		program = path.Base(program)
	}
	expression := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, ".*") + "$"
	matched, err := regexp.MatchString(expression, program)
	return err == nil && matched
}

func matchProgramRegex(expression string, value string) bool {
	if expression == "" {
		return true
	}
	matched, err := regexp.MatchString(expression, value)
	return err == nil && matched
}

func validateFormFillControl(value string, lang string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, part := range strings.Split(value, ";") {
		key, _, found := strings.Cut(part, "=")
		if !found || !formFillControlKeys[strings.TrimSpace(key)] {
			return fmt.Errorf(i18n.Translate(lang, "application:The control: %s is invalid, it should be like id=txtUser or type=Edit;index=1"), value)
		}
	}
	return nil
}

func matchFormFillUrl(formFill *object.FormFill, pageUrl string) bool {
	for _, pattern := range formFill.Urls {
		if MatchFormFillUrl(pattern, pageUrl) {
			return true
		}
	}
	return false
}

func matchFormFillTarget(formFill *object.FormFill, target *object.FormFillTarget) bool {
	if target.Program == "" {
		return matchFormFillUrl(formFill, target.Url)
	}
	for _, program := range formFill.Programs {
		if matchFormFillProgram(program, target) {
			return true
		}
	}
	return false
}

func matchFormFillProgram(program *object.FormFillProgram, target *object.FormFillTarget) bool {
	if program == nil {
		return false
	}
	if program.Publisher != "" && !strings.EqualFold(strings.TrimSpace(program.Publisher), strings.TrimSpace(target.Publisher)) {
		return false
	}
	if program.FileHash != "" && !strings.EqualFold(strings.TrimSpace(program.FileHash), strings.TrimSpace(target.FileHash)) {
		return false
	}
	if target.Launch {
		// starting the program with the account in its command line, before it has any window
		if !strings.Contains(program.LaunchArgs, "{username}") && !strings.Contains(program.LaunchArgs, "{password}") {
			return false
		}
		return matchProgramPath(program.Path, target.Program) || (program.LaunchPath != "" && matchProgramPath(program.LaunchPath, target.Program))
	}
	return matchProgramPath(program.Path, target.Program) && matchProgramRegex(program.WindowTitle, target.Title)
}

func validateFormFillProgram(program *object.FormFillProgram, lang string) error {
	if program == nil {
		return errors.New(i18n.Translate(lang, "application:The desktop program must have the path of its executable"))
	}
	if strings.TrimSpace(program.Path) == "" {
		return errors.New(i18n.Translate(lang, "application:The desktop program must have the path of its executable"))
	}
	for _, expression := range []string{program.WindowTitle, program.WindowClass} {
		if _, err := regexp.Compile(expression); err != nil {
			return fmt.Errorf(i18n.Translate(lang, "application:The regular expression: %s is invalid: %s"), expression, err.Error())
		}
	}
	for _, control := range []string{program.UsernameControl, program.PasswordControl, program.SubmitControl} {
		if err := validateFormFillControl(control, lang); err != nil {
			return err
		}
	}
	if !formFillFillMethods[program.FillMethod] {
		return fmt.Errorf(i18n.Translate(lang, "application:The fill method: %s is not supported"), program.FillMethod)
	}
	for _, match := range formFillKeyTokenRegex.FindAllStringSubmatch(program.KeySequence, -1) {
		if !formFillKeyTokens[match[1]] {
			return fmt.Errorf(i18n.Translate(lang, "application:The key: {%s} is not supported, use {username}, {password}, {TAB}, {ENTER} or {WAIT}"), match[1])
		}
	}
	return nil
}
