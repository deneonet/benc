import java.math.BigInteger;
import java.time.Instant;
import java.util.Arrays;

class BencValue {
  public static void main(String[] args) throws Exception {
    Root v = new Root();
    v.S = "A\0B";
    v.I = -7L;
    v.U = BigInteger.valueOf(300);
    v.Count = 201;
    v.Fixed = new byte[] {1, 2, 3};
    v.Numbers = Arrays.asList(-2, 5);
    v.Raw = new byte[] {0, (byte)255};
    Child ch = new Child(); ch.ID = -9L; ch.Text = "x"; ch.Scores = new byte[] {4, 5};
    v.Children.add(ch);
    v.Groups.add(new byte[] {0, 1});
    v.Named.value.put("n", -5);
    v.When = Instant.ofEpochSecond(123456789, 123);
    Child pointed = new Child(); pointed.ID = 3L; pointed.Text = "p"; pointed.Scores = new byte[] {6};
    v.P = pointed;
    Child mapped = new Child(); mapped.ID = 4L; mapped.Text = "z"; mapped.Scores = new byte[] {7};
    v.Dict.put("d", mapped);

    byte[] wire = new byte[v.size()];
    if (v.marshal(0, wire) != wire.length) throw new AssertionError("size mismatch");
    var decoded = Root.unmarshal(0, wire);
    if (decoded.newOffset() != wire.length || !decoded.value().When.equals(v.When)) throw new AssertionError("decode mismatch");
    byte[] again = new byte[decoded.value().size()];
    if (decoded.value().marshal(0, again) != again.length || !Arrays.equals(wire, again)) throw new AssertionError("roundtrip mismatch");
    System.out.println(java.util.HexFormat.of().formatHex(wire));
  }
}
