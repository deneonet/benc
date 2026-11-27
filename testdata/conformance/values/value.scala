import java.time.Instant

object BencValue:
  def main(args: Array[String]): Unit =
    val v = Root(
      S = "A\u0000B", I = -7L, U = 300L, Count = 201.toShort,
      Fixed = Array[Byte](1, 2, 3), Numbers = Seq(-2, 5), Raw = Array[Byte](0, 255.toByte),
      Children = Seq(Child(ID = -9, Text = "x", Scores = Array[Byte](4, 5))),
      Groups = Seq(Array[Byte](0, 1)),
      P = Some(Child(ID = 3, Text = "p", Scores = Array[Byte](6))),
      Dict = Map("d" -> Child(ID = 4, Text = "z", Scores = Array[Byte](7))),
      Named = Labels(Map("n" -> -5)),
      When = Instant.ofEpochSecond(123456789, 123)
    )
    val wire = new Array[Byte](Root.codec.size(v))
    assert(Root.codec.write(0, wire, v) == wire.length)
    val (end, copy) = Root.codec.read(0, wire).toOption.get
    assert(end == wire.length && copy.When == v.When)
    val again = new Array[Byte](Root.codec.size(copy))
    assert(Root.codec.write(0, again, copy) == again.length && wire.sameElements(again))
    println("WIRE=" + java.util.HexFormat.of().formatHex(wire))
