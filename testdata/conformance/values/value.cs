using System;
using System.Collections.Generic;

public static class BencValue {
  public static void Main() {
    var v = new Root();
    v.S = "A\0B"; v.I = -7; v.U = 300; v.Count = 201;
    v.Fixed = new byte[] {1, 2, 3}; v.Numbers = new List<int> {-2, 5};
    v.Raw = new byte[] {0, 255};
    v.Children = new List<Child> {new Child { ID = -9, Text = "x", Scores = new byte[] {4, 5} }};
    v.Groups = new List<byte[]> {new byte[] {0, 1}};
    v.Named.Value["n"] = -5;
    v.When = 123456789000000123;
    v.P = new Child { ID = 3, Text = "p", Scores = new byte[] {6} };
    v.Dict["d"] = new Child { ID = 4, Text = "z", Scores = new byte[] {7} };
    var wire = new byte[v.Size()]; Span<byte> writer = wire;
    v.Marshal(ref writer); if (!writer.IsEmpty) throw new Exception("size mismatch");
    ReadOnlySpan<byte> reader = wire; var copy = Root.Unmarshal(ref reader);
    if (!reader.IsEmpty || copy.When != v.When) throw new Exception("decode mismatch");
    var again = new byte[copy.Size()]; Span<byte> second = again;
    copy.Marshal(ref second); if (!second.IsEmpty || !wire.AsSpan().SequenceEqual(again)) throw new Exception("roundtrip mismatch");
    Console.WriteLine("WIRE=" + Convert.ToHexString(wire).ToLowerInvariant());
  }
}
