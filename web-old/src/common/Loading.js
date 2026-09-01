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

import React from "react";

// Palette: indigo → violet → cyan, evoking a modern AI aesthetic
const AI_COLORS = ["#6366F1", "#A855F7", "#06B6D4"];

/**
 * AiDots – the core animated indicator.
 * Used standalone and as antd's global Spin indicator (via ConfigProvider).
 *
 * size: "small" | "medium" (default) | "large"
 */
export const AiDots = ({size = "medium"}) => {
  const isSmall = size === "small";
  const isLarge = size === "large";

  const dotPx = isSmall ? 5 : isLarge ? 12 : 9;
  const gapPx = isSmall ? 5 : isLarge ? 10 : 8;
  const glowBlur = isSmall ? 0 : isLarge ? 12 : 8;

  return (
    <span style={{display: "inline-flex", alignItems: "center", gap: gapPx}}>
      {AI_COLORS.map((color, i) => (
        <span
          key={i}
          style={{
            display: "inline-block",
            width: dotPx,
            height: dotPx,
            borderRadius: "50%",
            background: color,
            animation: "casdoor-ai-bounce 1.4s ease-in-out infinite",
            animationDelay: `${i * 0.16}s`,
            boxShadow: glowBlur > 0 ? `0 0 ${glowBlur}px ${color}90` : "none",
          }}
        />
      ))}
    </span>
  );
};

/**
 * Loading – unified page / section / small loading component.
 *
 * Props:
 *   spinning  boolean   Whether to show loading (default: true). Returns null when false.
 *   tip       string    Optional label rendered below the dots.
 *   type      string    Layout preset:
 *                         "page"    – vertically centered in the viewport (calc(100vh - 120px))
 *                         "section" – padded block, suitable for card / content areas (default)
 *                         "small"   – no padding, tiny dots, no tip, for inline / dropdown use
 *   style     object    Extra styles applied to the outer wrapper (useful for absolute positioning).
 */
const Loading = ({spinning = true, tip, type = "section", style}) => {
  if (!spinning) {
    return null;
  }

  const isPage = type === "page";
  const isSmall = type === "small";

  const wrapperStyle = {
    display: "flex",
    flexDirection: "column",
    justifyContent: "center",
    alignItems: "center",
    ...(isPage && {width: "100%", height: "calc(100vh - 120px)"}),
    ...(type === "section" && {padding: "48px 0"}),
    ...style,
  };

  const tipStyle = {
    marginTop: 14,
    fontSize: 13,
    color: "#94A3B8",
    letterSpacing: "0.05em",
    fontWeight: 400,
    userSelect: "none",
  };

  return (
    <div style={wrapperStyle}>
      <AiDots size={isSmall ? "small" : isPage ? "large" : "medium"} />
      {tip && !isSmall && <div style={tipStyle}>{tip}</div>}
    </div>
  );
};

export default Loading;
