use traq_markdown_processor::{Processor, ProcessorOptions, ProcessorPreset};

fn processor(origin: &str) -> Processor {
    Processor::new(
        ProcessorPreset::TraQV1,
        ProcessorOptions {
            origin: origin.into(),
        },
    )
    .unwrap()
}

#[test]
fn preserves_all_notification_fixtures() {
    let processor = processor("https://q.example.test");
    let fixtures: Vec<serde_json::Value> = serde_json::from_str(include_str!(
        "../../../tests/fixtures/processing-notifications.json"
    ))
    .unwrap();
    assert_eq!(fixtures.len(), 787);
    for fixture in fixtures {
        let output = processor
            .process(fixture["source"].as_str().unwrap())
            .unwrap();
        assert_eq!(
            output.notification_text,
            fixture["notification"].as_str().unwrap(),
            "{}",
            fixture["name"]
        );
    }
}

#[test]
fn references_preserve_context_order_and_duplicates() {
    let processor = processor("");
    let id = "00000000-0000-0000-0000-000000000001";
    let reference = format!(r#"!{{"type":"user","id":"{id}","raw":"@alice"}}"#);
    let source = format!("{reference} !!{reference}!! `{reference}`");
    let result = processor.process(&source).unwrap();
    assert_eq!(result.references.mentions, [id, id]);
    assert!(result.references.group_mentions.is_empty());
    assert!(result.references.channel_links.is_empty());
    assert_eq!(
        result.notification_text,
        format!("@alice ██████ {reference}")
    );
    assert!(processor.process(&"x".repeat(65537)).is_err());
    assert_eq!(
        processor.process("**after**").unwrap().notification_text,
        "after"
    );
}

#[test]
fn origin_options_are_validated_and_isolated() {
    let plain = processor("");
    let configured = processor("https://q.example.test");
    let source = "https://q.example.test/files/00000000-0000-0000-0000-000000000001";
    assert_eq!(plain.process(source).unwrap().notification_text, source);
    assert_eq!(
        configured.process(source).unwrap().notification_text,
        "[添付ファイル]"
    );
    assert!(
        Processor::new(
            ProcessorPreset::TraQV1,
            ProcessorOptions {
                origin: "x".repeat(2049)
            }
        )
        .is_err()
    );
}
