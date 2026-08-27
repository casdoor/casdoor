import * as React from "react";
import CodeMirror from "@uiw/react-codemirror";
import {loadLanguage} from "@uiw/codemirror-extensions-langs";
import {material} from "@uiw/codemirror-theme-material";
import {useTheme} from "@/hooks/use-theme";

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  /** "javascript" | "json" | "html" | "css" | "sql" | "go" ... */
  language?: string;
  height?: number;
  readOnly?: boolean;
  className?: string;
}

/** CodeMirror 6 editor, replacing the antd-flavoured `common/Editor.js`. */
export function CodeEditor({value, onChange, language, height = 240, readOnly, className}: CodeEditorProps) {
  const {resolvedTheme} = useTheme();
  const extensions = React.useMemo(() => {
    if (!language) {
      return [];
    }
    const ext = loadLanguage(language as any);
    return ext ? [ext] : [];
  }, [language]);

  return (
    <div className={className}>
      <CodeMirror
        value={value ?? ""}
        height={`${height}px`}
        readOnly={readOnly}
        theme={resolvedTheme === "dark" ? material : "light"}
        extensions={extensions}
        onChange={(next) => onChange(next)}
        basicSetup={{lineNumbers: true, foldGutter: false, highlightActiveLine: !readOnly}}
      />
    </div>
  );
}

export default CodeEditor;
