use crate::buffers::IO;
use serde::Deserialize;
use std::cell::RefCell;
use traq_markdown_processor::{Processor, ProcessorOptions};

thread_local! { static PROCESSOR: RefCell<Option<Processor>> = const { RefCell::new(None) }; }

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Configuration {
    preset: String,
    options: ProcessorOptions,
}

#[unsafe(no_mangle)]
pub extern "C" fn configure_processor() -> u32 {
    IO.with_borrow_mut(|io| {
        let result = (|| -> Result<_, String> {
            let source = io.source().map_err(|error| error.to_string())?;
            let config: Configuration =
                serde_json::from_str(source).map_err(|error| error.to_string())?;
            let parser = traq_markdown_grammar::bindings::parser(&config.preset)
                .map_err(|error| error.to_string())?;
            let processor = Processor::new(parser, config.options)?;
            PROCESSOR.set(Some(processor));
            Ok(env!("MARKDOWN_BUILD_ID"))
        })();
        io.reply("configured", &result);
        io.output.len() as u32
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn process() -> u32 {
    IO.with_borrow_mut(|io| {
        let result = io
            .source()
            .map_err(|error| error.to_string())
            .and_then(|source| {
                PROCESSOR.with_borrow(|processor| {
                    processor
                        .as_ref()
                        .ok_or_else(|| "processor is not configured".to_owned())?
                        .process(source)
                })
            });
        io.reply("result", &result);
        io.output.len() as u32
    })
}
