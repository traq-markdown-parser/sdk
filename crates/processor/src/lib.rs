//! Distribution-owned processing presets. All consumers borrow one native AST.
#![forbid(unsafe_code)]

use markdown_extractor::Extractor;
use markdown_renderer::Renderer;
use serde::{Deserialize, Serialize};
use traq_markdown_grammar::Parser;
use traq_markdown_processing::{References, presets::traq::v1};

#[derive(Debug, Clone, Copy, Deserialize, Serialize)]
#[cfg_attr(feature = "contracts", derive(ts_rs::TS, schemars::JsonSchema))]
pub enum ProcessorPreset {
    #[serde(rename = "traq.v1")]
    TraQV1,
}

#[derive(Debug, Default, Deserialize, Serialize)]
#[cfg_attr(feature = "contracts", derive(ts_rs::TS, schemars::JsonSchema))]
#[serde(deny_unknown_fields)]
pub struct ProcessorOptions {
    /// Origin used to recognize traQ file/message URLs; empty leaves URLs as text.
    #[serde(default)]
    pub origin: String,
}

#[derive(Debug, PartialEq, Serialize)]
#[cfg_attr(feature = "contracts", derive(ts_rs::TS, schemars::JsonSchema))]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ProcessOutput {
    pub notification_text: String,
    pub references: References,
}

/// Owns the Rust implementations, without external document handles or AST copies.
pub struct Processor {
    parser: Parser,
    renderer: Renderer,
    extractor: Extractor<References>,
}

impl Processor {
    pub fn new(preset: ProcessorPreset, options: ProcessorOptions) -> Result<Self, &'static str> {
        match preset {
            ProcessorPreset::TraQV1 => Ok(Self {
                parser: traq_markdown_grammar::presets::traq::v1::parser(),
                renderer: Renderer::new(&v1::notification::preset(&options.origin)?),
                extractor: Extractor::new(&v1::references::preset()?),
            }),
        }
    }

    pub fn process(&self, source: &str) -> Result<ProcessOutput, String> {
        let document = self
            .parser
            .parse(source)
            .map_err(|error| error.to_string())?;

        let notification_text = self
            .renderer
            .render(&document)?
            .split_whitespace()
            .collect::<Vec<_>>()
            .join(" ");

        let references = self.extractor.extract(&document)?;
        Ok(ProcessOutput {
            notification_text,
            references,
        })
    }
}
