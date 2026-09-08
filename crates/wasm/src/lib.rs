//! Instance-local composition and parsing ABI. No WASI or host callbacks.
mod buffers;
mod contract;
mod grammar;
#[macro_use]
mod node_types;
mod nodes;
use buffers::IO;
use markdown_traq::{ParseError, Parser};

#[unsafe(no_mangle)]
pub extern "C" fn abi_version() -> u32 {
    2
}
#[unsafe(no_mangle)]
pub extern "C" fn ast_version() -> u32 {
    4
}

/// mode: 0=document, 1=inline. Handles belong to this Wasm instance.
#[unsafe(no_mangle)]
pub extern "C" fn parse(handle: u32, mode: u32) -> u32 {
    IO.with_borrow_mut(|io| {
        let result = io.source().and_then(|source| {
            let grammar = grammar::get(handle).ok_or(ParseError::InternalError)?;
            let parser = Parser::new(&grammar);
            match mode {
                0 => parser.parse(source),
                1 => parser.parse_inline(source),
                _ => Err(ParseError::InternalError),
            }
        });
        io.reply_document(&result);
        io.output.len() as u32
    })
}
