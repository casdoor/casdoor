#!/usr/bin/env python3
import json

def load_json(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        return json.load(f)

def find_untranslated(lang):
    base_frontend = "/home/runner/work/casdoor/casdoor/web/src/locales"
    base_backend = "/home/runner/work/casdoor/casdoor/i18n/locales"
    
    en_frontend = load_json(f"{base_frontend}/en/data.json")
    en_backend = load_json(f"{base_backend}/en/data.json")
    
    lang_frontend = load_json(f"{base_frontend}/{lang}/data.json")
    lang_backend = load_json(f"{base_backend}/{lang}/data.json")
    
    print(f"\n=== {lang.upper()} Untranslated Entries ===")
    print("\nFrontend:")
    def check_dict(en_dict, lang_dict, prefix=""):
        for key, en_val in en_dict.items():
            full_key = f"{prefix}.{key}" if prefix else key
            if isinstance(en_val, dict):
                if key in lang_dict and isinstance(lang_dict[key], dict):
                    check_dict(en_val, lang_dict[key], full_key)
            else:
                lang_val = lang_dict.get(key) if not prefix else lang_dict.get(key)
                if lang_val == en_val:
                    print(f"  {full_key}: '{en_val}'")
    
    check_dict(en_frontend, lang_frontend)
    
    print("\nBackend:")
    check_dict(en_backend, lang_backend)

# Check first language
find_untranslated('fr')
