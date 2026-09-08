use super::buffers::IO;
use markdown_traq::{Grammar, bindings, engine::BuildError};
use std::cell::RefCell;

thread_local! { static GRAMMAR: RefCell<Option<Grammar>> = const { RefCell::new(None) }; }

pub fn get() -> Option<Grammar> {
    GRAMMAR.with_borrow(Clone::clone)
}

fn select(name: &str) -> Result<Grammar, BuildError> {
    let catalog = bindings::bundled();
    let mut value = &catalog.exports["presets"];
    for part in name.split('.') {
        value = &value[part];
    }
    let composition = value
        .as_u64()
        .and_then(|index| catalog.preset_composition(index as usize))
        .ok_or_else(|| BuildError::InvalidDefinition {
            reason: format!("unknown preset: {name}"),
        })?;
    catalog.build(&composition)
}

/// One Rust-owned grammar per instance. Hosts select a preset, not its recipe.
#[unsafe(no_mangle)]
pub extern "C" fn configure() -> u32 {
    IO.with_borrow_mut(|io| {
        let result = io
            .source()
            .map_err(|_| BuildError::InvalidDefinition {
                reason: "invalid preset encoding".into(),
            })
            .and_then(select)
            .map(|grammar| {
                GRAMMAR.set(Some(grammar));
                env!("MARKDOWN_BUILD_ID")
            });
        io.reply("configured", &result);
        io.output.len() as u32
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn selection_uses_rust_exports_and_rejects_unknown_presets() {
        assert!(select("commonmark").is_ok());
        assert!(select("traq.v1").is_ok());
        for name in ["", "traq", "traq.v2", "commonmark.extra"] {
            assert!(select(name).is_err());
        }
    }
}
