use markdown_parser::{Parser, bindings::Composition};
use markdown_traq::{bindings::bundled, presets};

#[test]
fn bundled_compositions_preserve_presets_and_validate_untrusted_recipes() {
    let catalog = bundled();
    for (index, expected) in [presets::commonmark::grammar(), presets::traq::v1::grammar()]
        .into_iter()
        .enumerate()
    {
        assert_eq!(
            catalog.describe()["presets"][index]["description"],
            expected.describe()
        );
        let recipe = catalog.preset_composition(index).unwrap();
        let recipe: Composition =
            serde_json::from_str(&serde_json::to_string(&recipe).unwrap()).unwrap();
        let grammar = catalog.build(&recipe).unwrap();
        assert_eq!(grammar.describe(), expected.describe());
        for source in [
            "**text** <i>x</i>",
            "$x$ ==x== :stamp:",
            "- item\n\t- child",
            "!{\"type\":\"user\",\"id\":\"u\",\"raw\":\"@u\"}",
        ] {
            assert_eq!(
                Parser::new(&grammar).parse(source).unwrap(),
                Parser::new(expected).parse(source).unwrap()
            );
        }
        let mut incomplete = recipe.clone();
        incomplete.order.pop();
        assert!(catalog.build(&incomplete).is_err());
        let mut duplicate = recipe.clone();
        duplicate.order[1] = duplicate.order[0];
        assert!(catalog.build(&duplicate).is_err());
        let mut unknown = recipe;
        unknown.plugins[0].rules[0] = usize::MAX;
        assert!(catalog.build(&unknown).is_err());
    }
}
