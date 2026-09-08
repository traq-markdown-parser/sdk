use std::sync::LazyLock;
static CONTRACT: LazyLock<String> = LazyLock::new(|| {
    serde_json::json!({
        "abiVersion": 2, "astVersion": 4,
        "catalog": markdown_traq::bindings::bundled().describe(),
        "limits": {
            "inputBytes": super::buffers::MAX_INPUT,
            "outputBytes": super::buffers::MAX_OUTPUT,
            "memoryBytes": 33_554_432,
            "grammars": super::grammar::MAX_GRAMMARS
        }
    })
    .to_string()
});
#[unsafe(no_mangle)]
pub extern "C" fn contract_ptr() -> u32 {
    CONTRACT.as_ptr() as u32
}
#[unsafe(no_mangle)]
pub extern "C" fn contract_len() -> u32 {
    CONTRACT.len() as u32
}
