#[path = "../model_benc.rs"]
mod model;

use model::{Child, Root};

fn main() {
    let mut v = Root::default();
    v.S = "A\0B".into();
    v.I = -7;
    v.U = 300;
    v.Count = 201;
    v.Fixed = [1, 2, 3];
    v.Numbers = [-2, 5];
    v.Raw = vec![0, 255];
    v.Children = vec![Child { ID: -9, Text: "x".into(), Scores: vec![4, 5] }];
    v.Groups = vec![vec![0, 1]];
    v.Named.0.insert("n".into(), -5);
    v.When = chrono::DateTime::from_timestamp_nanos(123456789000000123);
    v.P = Some(Child { ID: 3, Text: "p".into(), Scores: vec![6] });
    v.Dict.insert("d".into(), Child { ID: 4, Text: "z".into(), Scores: vec![7] });
    let mut wire = vec![0; v.size()];
    let mut writer = wire.as_mut_slice();
    v.marshal(&mut writer).unwrap();
    assert!(writer.is_empty());
    let mut reader = wire.as_slice();
    let copy = Root::unmarshal(&mut reader).unwrap();
    assert!(reader.is_empty());
    assert_eq!(copy, v);
    println!("WIRE={}", wire.iter().map(|b| format!("{b:02x}")).collect::<String>());
}
