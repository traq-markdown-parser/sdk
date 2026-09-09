import type { Options as CommonOptions } from "@traq-markdown-parser/commonmark/renderer";
import type { Options as GenericOptions } from "@traq-markdown-parser/commonmark/generic/renderer";
import type { Options as TrapOptions } from "@traq-markdown-parser/trap-extension/renderer";
import { plugin as common } from "@traq-markdown-parser/commonmark/renderer";
import { plugin as generic } from "@traq-markdown-parser/commonmark/generic/renderer";
import { plugin as trap } from "@traq-markdown-parser/trap-extension/renderer";
import { PresetBuilder } from "@traq-markdown-parser/core/renderer";
import { createHighlightFunc } from "@traq-markdown-parser/commonmark/highlight";
import type { Document } from "@traq-markdown-parser/core/renderer";
import { renderer } from "@traq-markdown-parser/core/renderer";
import { prepareMessage } from "./embeddings.js";
import { configurePreview } from "./preview.js";
import { math } from "@traq-markdown-parser/commonmark/generic/math";
export { embeddingFromUrl } from "./embeddings.js";
export type { Embedding } from "./embeddings.js";
import imageDomains from "./image-domains.js";

export type Options = CommonOptions & GenericOptions & TrapOptions;

const highlight = createHighlightFunc("traq-code traq-lang");

const validateImage = (value: string) => {
  try {
    const url = new URL(value);
    return url.protocol === "https:" && imageDomains.includes(url.hostname);
  } catch {
    return false;
  }
};

function previewOptions(options: Options) {
  const customMath = options.math;

  return {
    ...options,
    math: customMath
      ? (tex: string) => customMath(tex, false)
      : (tex: string) =>
          math(tex, false, {
            maxSize: 1,
            macros: {
              "\\Huge": "",
              "\\huge": "",
              "\\LARGE": "",
              "\\Large": "",
              "\\large": "",
            },
          }),
  };
}

function build(options: Options = {}, preview = false) {
  const commonPlugin = common({
    breaks: true,
    highlight,
    validateImage,
    linkAttributes: { target: "_blank", rel: "nofollow noopener noreferrer" },
    ...options,
  });
  const genericPlugin = generic(preview ? previewOptions(options) : options);

  const trapPlugin = trap(options);

  if (preview)
    configurePreview(commonPlugin, genericPlugin, trapPlugin, options);

  return new PresetBuilder()
    .add(commonPlugin)
    .add(genericPlugin)
    .add(trapPlugin)
    .build();
}

export function html(options?: Options) {
  return build(options);
}

/** Render a full Rust-parsed document as message content or a one-line preview. */
export function messageRenderer({
  origin,
  ...options
}: Options & { origin: string }) {
  const embeddingOrigin = new URL(origin).origin;
  const full = renderer(build(options)),
    preview = renderer(build(options, true));

  function render(document: Document, inline: boolean) {
    const prepared = prepareMessage(document, embeddingOrigin, inline);
    const renderedText = inline
      ? prepared.document.children
          .map((node) =>
            preview.renderInline({ ...prepared.document, children: [node] }),
          )
          .join(" ")
      : full.render(prepared.document);
    return {
      rawText: document.source,
      renderedText,
      embeddings: prepared.embeddings,
    };
  }

  return Object.freeze({
    render: (document: Document) => render(document, false),
    renderInline: (document: Document) => render(document, true),
  });
}
