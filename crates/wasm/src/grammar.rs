use super::buffers::IO;
use markdown_traq::{
    Grammar,
    bindings::{self, Composition},
    engine::BuildError,
};
use serde::Serialize;
use std::{cell::RefCell, collections::HashMap};

pub const MAX_GRAMMARS: usize = 256;
#[derive(Default)]
struct Grammars {
    next: u32,
    items: HashMap<u32, Grammar>,
}
thread_local! { static GRAMMARS: RefCell<Grammars> = RefCell::default(); }
#[derive(Serialize)]
struct Built {
    handle: u32,
    description: String,
}
fn invalid(reason: &str) -> BuildError {
    BuildError::InvalidDefinition {
        reason: reason.into(),
    }
}
pub fn get(handle: u32) -> Option<Grammar> {
    GRAMMARS.with_borrow(|state| state.items.get(&handle).cloned())
}
fn build(source: &str) -> Result<Built, BuildError> {
    let spec: Composition =
        serde_json::from_str(source).map_err(|_| invalid("invalid composition JSON"))?;
    GRAMMARS.with_borrow_mut(|state| {
        if state.items.len() >= MAX_GRAMMARS {
            return Err(invalid("live grammar limit exceeded"));
        }
        let handle = state
            .next
            .checked_add(1)
            .ok_or_else(|| invalid("grammar handles exhausted"))?;
        let grammar = bindings::bundled().build(&spec)?;
        let value = Built {
            handle,
            description: grammar.describe(),
        };
        // Never reuse identifiers: a released handle cannot address a later grammar.
        state.next = handle;
        state.items.insert(handle, grammar);
        Ok(value)
    })
}
#[unsafe(no_mangle)]
pub extern "C" fn grammar_build() -> u32 {
    IO.with_borrow_mut(|io| {
        let result = io
            .source()
            .map_err(|_| invalid("invalid composition encoding"))
            .and_then(build);
        if !io.reply("grammar", &result)
            && let Ok(value) = result
        {
            grammar_drop(value.handle);
        }
        io.output.len() as u32
    })
}
#[unsafe(no_mangle)]
pub extern "C" fn grammar_drop(handle: u32) -> u32 {
    GRAMMARS.with_borrow_mut(|state| u32::from(state.items.remove(&handle).is_some()))
}
#[unsafe(no_mangle)]
pub extern "C" fn grammar_count() -> u32 {
    GRAMMARS.with_borrow(|state| state.items.len() as u32)
}
