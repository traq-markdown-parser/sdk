use traq_markdown_grammar::{Parser, presets, syntax::extensions::math};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = "**hello** :stamp: $x$";
    let parser = presets::traq::v1::parser();
    println!("{:#?}", parser.parse(source)?);

    let mut builder = presets::traq::v1::grammar().to_builder();
    builder.remove(math::plugin())?;
    let grammar = builder.build()?;
    let without_math = Parser::new(&grammar);
    drop(grammar);
    println!("{:#?}", without_math.parse_inline("$x$")?);
    let processor = traq_markdown_processor::Processor::new(
        traq_markdown_processor::ProcessorPreset::TraQV1,
        traq_markdown_processor::ProcessorOptions {
            origin: "https://q.example.test".into(),
        },
    )?;
    println!("{:#?}", processor.process("**hello** !!secret!!")?);
    Ok(())
}
