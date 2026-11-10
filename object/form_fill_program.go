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

package object

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/i18n"
)

// FormFillProgram is a desktop (C/S) program whose sign-in window the desktop autofill client fills
type FormFillProgram struct {
	Name string `json:"name"`
	// the executable: a file name like "u8client.exe", or a full path where * matches anything, like
	// "C:\Program Files\UFIDA\*\u8client.exe"
	Path string `json:"path"`
	// the subject CN of the Authenticode certificate that the executable must be signed with, and its SHA-256, both
	// optional, so that another program with the same name doesn't get the password
	Publisher string `json:"publisher"`
	FileHash  string `json:"fileHash"`
	// regular expressions of the sign-in window's title and its window class, any window of the program when empty
	WindowTitle string `json:"windowTitle"`
	WindowClass string `json:"windowClass"`
	// the controls, like "id=txtUser", "name=Password;type=Edit" or "type=Edit;index=1", found automatically when empty
	UsernameControl string `json:"usernameControl"`
	PasswordControl string `json:"passwordControl"`
	SubmitControl   string `json:"submitControl"`
	// "" fills the controls (UI Automation, then Win32 messages), "Keyboard" types the key sequence into the window
	// for the programs whose controls can't be seen, like the self-drawn ones
	FillMethod  string `json:"fillMethod"`
	KeySequence string `json:"keySequence"`
	// how the client starts the program from its menu or the portal, the arguments may have {username} and {password}
	// for the programs that take the account in their command line
	LaunchPath string `json:"launchPath"`
	LaunchArgs string `json:"launchArgs"`
}

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

func (program *FormFillProgram) match(target *FormFillTarget) bool {
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

func (program *FormFillProgram) validate(lang string) error {
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
