//! Catalog composition belongs to this distribution, not the parser core.
pub use markdown_parser::bindings::*;

use std::sync::LazyLock;

pub fn bundled() -> &'static Catalog {
    static CATALOG: LazyLock<Catalog> = LazyLock::new(crate::presets::exports::catalog);
    &CATALOG
}
