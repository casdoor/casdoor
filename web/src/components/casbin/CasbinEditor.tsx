import * as React from "react";
import i18next from "i18next";
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {CodeEditor} from "@/components/common/CodeEditor";
import * as Setting from "@/lib/setting";

const EDITOR_ORIGIN = "https://editor.casbin.org";

interface IframeEditorHandle {
  getModelText: () => void;
  updateModelText: (modelText: string) => void;
}

interface IframeEditorProps {
  initialModelText: string;
  onModelTextChange: (modelText: string) => void;
}

/**
 * The visual Casbin model editor, embedded from editor.casbin.org. Ported from
 * web/src/IframeEditor.js — the postMessage protocol ("iframeReady",
 * "initializeModel", "getModelText", "updateModelText", "modelUpdate") is
 * unchanged, so the same editor build keeps working.
 */
const IframeEditor = React.forwardRef<IframeEditorHandle, IframeEditorProps>(
  ({initialModelText, onModelTextChange}, ref) => {
    const iframeRef = React.useRef<HTMLIFrameElement>(null);
    const [iframeReady, setIframeReady] = React.useState(false);
    const currentLang = localStorage.getItem("language") || "en";

    React.useEffect(() => {
      const handleMessage = (event: MessageEvent) => {
        if (event.origin !== EDITOR_ORIGIN) {
          return;
        }
        if (event.data.type === "modelUpdate") {
          onModelTextChange(event.data.modelText);
        } else if (event.data.type === "iframeReady") {
          setIframeReady(true);
          if (initialModelText && iframeRef.current?.contentWindow) {
            iframeRef.current.contentWindow.postMessage({
              type: "initializeModel",
              modelText: initialModelText,
              lang: currentLang,
            }, "*");
          }
        }
      };

      window.addEventListener("message", handleMessage);
      return () => window.removeEventListener("message", handleMessage);
    }, [onModelTextChange, initialModelText, currentLang]);

    React.useImperativeHandle(ref, () => ({
      getModelText: () => {
        iframeRef.current?.contentWindow?.postMessage({type: "getModelText"}, "*");
      },
      updateModelText: (newModelText: string) => {
        if (iframeReady) {
          iframeRef.current?.contentWindow?.postMessage({type: "updateModelText", modelText: newModelText}, "*");
        }
      },
    }));

    return (
      <iframe
        ref={iframeRef}
        src={`${EDITOR_ORIGIN}/model-editor?lang=${currentLang}`}
        width="100%"
        height="500px"
        className="rounded-md border"
        title="Casbin Model Editor"
      />
    );
  },
);
IframeEditor.displayName = "IframeEditor";

interface CasbinEditorProps {
  model: any;
  onModelTextChange: (modelText: string) => void;
}

/**
 * "Basic" (plain text) and "Advanced" (editor.casbin.org) views of a model's
 * text, ported from web/src/CasbinEditor.js. Switching tabs pulls the current
 * text out of the iframe first so neither side loses an edit.
 */
export function CasbinEditor({model, onModelTextChange}: CasbinEditorProps) {
  const [activeKey, setActiveKey] = React.useState("advanced");
  const iframeRef = React.useRef<IframeEditorHandle>(null);
  const [localModelText, setLocalModelText] = React.useState(model.modelText);

  const isBuiltIn = Setting.builtInObject(model);

  const handleModelTextChange = React.useCallback((newModelText: string) => {
    if (!Setting.builtInObject(model)) {
      setLocalModelText(newModelText);
      onModelTextChange(newModelText);
    }
  }, [model, onModelTextChange]);

  const syncModelText = React.useCallback(() => {
    return new Promise<void>((resolve) => {
      if (activeKey === "advanced" && iframeRef.current) {
        const handleSyncMessage = (event: MessageEvent) => {
          if (event.data.type === "modelUpdate") {
            window.removeEventListener("message", handleSyncMessage);
            handleModelTextChange(event.data.modelText);
            resolve();
          }
        };
        window.addEventListener("message", handleSyncMessage);
        iframeRef.current.getModelText();
      } else {
        resolve();
      }
    });
  }, [activeKey, handleModelTextChange]);

  const handleTabChange = (key: string) => {
    syncModelText().then(() => {
      setActiveKey(key);
      if (key === "advanced") {
        iframeRef.current?.updateModelText(localModelText);
      }
    });
  };

  React.useEffect(() => {
    setLocalModelText(model.modelText);
  }, [model.modelText]);

  return (
    <div className="space-y-2">
      <Tabs value={activeKey} onValueChange={handleTabChange}>
        <TabsList>
          <TabsTrigger value="basic">{i18next.t("model:Basic Editor")}</TabsTrigger>
          <TabsTrigger value="advanced">{i18next.t("model:Advanced Editor")}</TabsTrigger>
        </TabsList>
      </Tabs>
      {activeKey === "advanced" ? (
        <IframeEditor
          ref={iframeRef}
          initialModelText={localModelText}
          onModelTextChange={handleModelTextChange}
        />
      ) : (
        <CodeEditor
          height={500}
          value={localModelText ?? ""}
          readOnly={isBuiltIn}
          onChange={handleModelTextChange}
        />
      )}
    </div>
  );
}
