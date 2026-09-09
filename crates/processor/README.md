# Native processing

Resolve the stored grammar version with
`traq_markdown_grammar::bindings::parser(version)?`, then pass that parser to
`Processor::new(parser, ProcessorOptions { origin })`. The processor parses once
and lends the AST to notification rendering, reference extraction, message metadata
and embedding planning. The processing composition has no grammar version of its own.

A caller can also supply a parser built from a custom grammar. Processing and
rendering evolve with the package version; grammar selection remains explicit.
No codec or serialized intermediate AST is used. See the root README for host
usage and `examples/rust` for a native consumer.
