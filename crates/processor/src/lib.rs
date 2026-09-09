//! Distribution-owned processing presets. All consumers borrow one native AST.
#![forbid(unsafe_code)]

use markdown_extractor::Extractor;
use markdown_renderer::Renderer;
use serde::{Deserialize, Serialize};
use traq_markdown_grammar::Parser;
use traq_markdown_processing::{References, presets::traq};

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
    pub plain_text: String,
    pub attachments: Vec<String>,
    pub citations: Vec<String>,
    pub references: References,
    pub embedding: traq::embedding::EmbeddingPlan,
}

/// Owns the Rust implementations, without external document handles or AST copies.
pub struct Processor {
    parser: Parser,
    renderer: Renderer,
    extractor: Extractor<References>,
    message: traq::message::Processor,
}

impl Processor {
    pub fn new(parser: Parser, options: ProcessorOptions) -> Result<Self, &'static str> {
        Ok(Self {
            parser,
            renderer: Renderer::new(&traq::notification::preset(&options.origin)?),
            extractor: Extractor::new(&traq::references::preset()?),
            message: traq::message::Processor::new(&options.origin),
        })
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
        let message = self.message.process(&document)?;
        Ok(ProcessOutput {
            notification_text,
            plain_text: message.plain_text,
            attachments: message.attachments,
            citations: message.citations,
            references,
            embedding: traq::embedding::plan(&document).map_err(str::to_owned)?,
        })
    }
}
