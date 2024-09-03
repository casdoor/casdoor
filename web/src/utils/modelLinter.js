import {newModel} from "casbin";
import "codemirror/lib/codemirror.css";
import "codemirror/addon/lint/lint.css";
import "codemirror/addon/lint/lint";

export const checkModelSyntax = (modelText) => {
  try {
    const model = newModel(modelText);
    if (!model.model.get("r") || !model.model.get("p") || !model.model.get("e")) {
      throw new Error("Model is missing one or more required sections (r, p, or e)");
    }
    return null;
  } catch (e) {
    return e.message;
  }
};

export const createLinter = (CodeMirror) => {
  CodeMirror.registerHelper("lint", "properties", (text) => {
    const error = checkModelSyntax(text);
    if (error) {
      const lineMatch = error.match(/line (\d+)/);
      if (lineMatch) {
        const lineNumber = parseInt(lineMatch[1], 10) - 1;
        return [{
          from: CodeMirror.Pos(lineNumber, 0),
          to: CodeMirror.Pos(lineNumber, text.split("\n")[lineNumber].length),
          message: error,
          severity: "error",
        }];
      } else {
        return [{
          from: CodeMirror.Pos(0, 0),
          to: CodeMirror.Pos(text.split("\n").length - 1),
          message: error,
          severity: "error",
        }];
      }
    }
    return [];
  });
};
