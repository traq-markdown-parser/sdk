import type { ParseError, BuildError } from "./types.js";
export class MarkdownParseError extends Error {
  readonly detail: ParseError;
  constructor(detail: ParseError) {
    super(detail.code);
    this.name = "MarkdownParseError";
    this.detail = detail;
  }
}
export class GrammarBuildError extends Error {
  readonly detail: BuildError;
  constructor(detail: BuildError) {
    super(
      detail.code +
        ": " +
        ("reason" in detail
          ? detail.reason
          : "element" in detail
            ? detail.element
            : [detail.scope, detail.name].join(" / ")),
    );
    this.name = "GrammarBuildError";
    this.detail = detail;
  }
}
