# Translation Workflow for Casdoor

## Current Status

After the i18n improvements in this PR:
- ✅ All tooltips have proper English descriptions
- ✅ All JSON files are structurally correct
- ✅ Crowdin is configured for both frontend and backend
- 🔄 ~2,500 strings need professional translation across 26 languages

## Workflow for Translators

### 1. Using Crowdin (Recommended)

Crowdin is already configured for this project. To translate:

1. **Sync English source to Crowdin:**
   ```bash
   cd web/
   yarn crowdin:sync
   ```

2. **Translate on Crowdin platform:**
   - Visit: https://crowdin.com/project/casdoor-site
   - Select your language
   - Translate untranslated strings
   - Crowdin provides context and suggestions

3. **Download translations:**
   ```bash
   cd web/
   yarn crowdin:sync  # Downloads and uploads
   ```

### 2. Manual Translation (Not Recommended)

If translating manually:

1. **Edit the JSON file for your language:**
   - Frontend: `web/src/locales/{lang_code}/data.json`
   - Backend: `i18n/locales/{lang_code}/data.json`

2. **Find untranslated strings:**
   - Look for values that match the English file
   - Tooltips ending in " - Tooltip" need translation
   - Simple field names need localization

3. **Translate:**
   - Keep JSON structure intact
   - Preserve placeholders like `%s`, `%d`, `%w`
   - Maintain HTML entities like `\"`
   - Keep formatting consistent

4. **Validate:**
   ```bash
   python3 -m json.tool your-file.json
   ```

### 3. Language Codes

The project supports 26 languages:
- ar (Arabic), az (Azerbaijani), cs (Czech), de (German)
- en (English - reference), es (Spanish), fa (Persian), fi (Finnish)
- fr (French), he (Hebrew), id (Indonesian), it (Italian)
- ja (Japanese), kk (Kazakh), ko (Korean), ms (Malay)
- nl (Dutch), pl (Polish), pt (Portuguese), ru (Russian)
- sk (Slovak), sv (Swedish), tr (Turkish), uk (Ukrainian)
- vi (Vietnamese), zh (Chinese Simplified)

## Priority Languages

Based on untranslated string counts:
1. **nl (Dutch)**: 202 strings
2. **az (Azerbaijani)**: 193 strings
3. **id (Indonesian)**: 188 strings
4. **it (Italian)**: 182 strings
5. **ms (Malay)**: 172 strings

## Translation Guidelines

### Do's
✅ Use native, natural language
✅ Maintain technical terms consistently
✅ Preserve placeholders and formatting
✅ Test in the application UI
✅ Review context from English descriptions

### Don'ts
❌ Don't translate technical terms that are universal (e.g., "OAuth", "SAML", "JWT")
❌ Don't remove or change placeholders
❌ Don't break JSON syntax
❌ Don't translate URLs or code examples
❌ Don't use machine translation without review

## Special Cases

### Tooltips
Format: `"Field Name - Tooltip": "Description of what this field does"`

Example:
```json
"Password obfuscator - Tooltip": "Method used to obfuscate passwords when sent from frontend to backend"
```

### Placeholders
Keep format specifiers intact:
- `%s` - string
- `%d` - number
- `%w` - error object

Example:
```json
"The application: %s does not exist": "La aplicación: %s no existe"
```

### HTML Entities
Preserve escaped characters:
```json
"Please enable \\\"Signin session\\\" first": "请先启用\\\"登录会话\\\""
```

## Testing Translations

1. **Build the frontend:**
   ```bash
   cd web/
   yarn build
   ```

2. **Change language in UI:**
   - Run the application
   - Switch to your language in settings
   - Check for:
     - Proper display of text
     - No overflow or truncation
     - Correct RTL for Arabic/Hebrew
     - Context appropriateness

3. **Verify backend:**
   ```bash
   go run main.go
   # Test API error messages in your language
   ```

## Questions?

- GitHub Issues: https://github.com/casdoor/casdoor/issues
- Crowdin Platform: https://crowdin.com/project/casdoor-site
- Community: Join the Discord (link in README)

---

**Note:** This PR has fixed structural issues. The remaining translation work is regular i18n maintenance that should go through Crowdin for quality assurance.
