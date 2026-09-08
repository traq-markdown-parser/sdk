# Native processing

`Processor::new(ProcessorPreset::TraQV1, ProcessorOptions { origin })` creates
one parser and the notification and reference consumers. `process(source)`
parses once, then lends its Document to both consumers and returns ProcessOutput.

This crate owns the SDK's fixed processing composition. Rendering and extraction
implementations live in the core, commonmark and trap repositories. No codec or
serialized intermediate AST is used. See the root README for host usage and
`examples/rust` for a native consumer.
