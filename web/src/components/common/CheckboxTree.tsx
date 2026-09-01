import * as React from "react";
import {ChevronRight} from "lucide-react";
import {Checkbox} from "@/components/ui/checkbox";
import {cn} from "@/lib/utils";

export interface TreeNode {
  key: string;
  title: React.ReactNode;
  children?: TreeNode[];
}

interface CheckboxTreeProps {
  nodes: TreeNode[];
  checkedKeys: string[];
  onCheck: (checkedKeys: string[]) => void;
  disabled?: boolean;
  /** keys of the branches that start expanded */
  defaultExpandedKeys?: string[];
  className?: string;
}

function collectKeys(node: TreeNode, into: string[] = []): string[] {
  into.push(node.key);
  node.children?.forEach((child) => collectKeys(child, into));
  return into;
}

/**
 * The shadcn stand-in for antd's `<Tree checkable />`, used by the organization
 * navbar / widget item pickers. It keeps antd's check semantics, which Casdoor's
 * stored `navItems` depend on: checking a branch checks its whole subtree, and a
 * branch is checked only while every one of its children is.
 */
export function CheckboxTree({
  nodes,
  checkedKeys,
  onCheck,
  disabled,
  defaultExpandedKeys = [],
  className,
}: CheckboxTreeProps) {
  const [expanded, setExpanded] = React.useState<string[]>(defaultExpandedKeys);
  const checked = React.useMemo(() => new Set(checkedKeys), [checkedKeys]);

  const setChecked = (node: TreeNode, next: boolean) => {
    const subtree = collectKeys(node);
    const result = new Set(checked);
    subtree.forEach((key) => (next ? result.add(key) : result.delete(key)));

    // a parent is only checked while all of its children are, so re-derive the
    // branches bottom-up after the subtree changed
    const reconcile = (candidate: TreeNode): boolean => {
      if (!candidate.children || candidate.children.length === 0) {
        return result.has(candidate.key);
      }
      const allChecked = candidate.children.map(reconcile).every(Boolean);
      if (allChecked) {
        result.add(candidate.key);
      } else {
        result.delete(candidate.key);
      }
      return allChecked;
    };
    nodes.forEach(reconcile);

    onCheck([...result]);
  };

  const renderNode = (node: TreeNode, depth: number): React.ReactNode => {
    const hasChildren = !!node.children?.length;
    const isExpanded = expanded.includes(node.key);
    const isChecked = checked.has(node.key);
    const descendants = hasChildren ? collectKeys(node).slice(1) : [];
    const isIndeterminate = hasChildren && !isChecked && descendants.some((key) => checked.has(key));

    return (
      <li key={node.key}>
        <div className="flex items-center gap-1.5 py-1" style={{paddingLeft: `${depth * 20}px`}}>
          {hasChildren ? (
            <button
              type="button"
              aria-label={isExpanded ? "Collapse" : "Expand"}
              className="rounded p-0.5 text-muted-foreground hover:bg-accent"
              onClick={() =>
                setExpanded((prev) =>
                  prev.includes(node.key) ? prev.filter((key) => key !== node.key) : [...prev, node.key],
                )
              }
            >
              <ChevronRight className={cn("h-3.5 w-3.5 transition-transform", isExpanded && "rotate-90")} />
            </button>
          ) : (
            <span className="w-[1.375rem]" />
          )}
          <Checkbox
            id={`tree-${node.key}`}
            disabled={disabled}
            checked={isIndeterminate ? "indeterminate" : isChecked}
            onCheckedChange={(next) => setChecked(node, next === true)}
          />
          <label
            htmlFor={`tree-${node.key}`}
            className={cn("cursor-pointer select-none text-sm", disabled && "cursor-not-allowed opacity-60")}
          >
            {node.title}
          </label>
        </div>
        {hasChildren && isExpanded ? (
          <ul>{node.children!.map((child) => renderNode(child, depth + 1))}</ul>
        ) : null}
      </li>
    );
  };

  return (
    <ul className={cn("rounded-lg border p-2", className)}>{nodes.map((node) => renderNode(node, 0))}</ul>
  );
}
