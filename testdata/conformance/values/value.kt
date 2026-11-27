import java.time.Instant

fun main() {
    val v = Root()
    v.S = "A\u0000B"
    v.I = -7L
    v.U = 300uL
    v.Count = 201u
    v.Fixed = byteArrayOf(1, 2, 3)
    v.Numbers = mutableListOf(-2, 5)
    v.Raw = byteArrayOf(0, 255.toByte())
    v.Children = mutableListOf(Child().apply { ID = -9L; Text = "x"; Scores = byteArrayOf(4, 5) })
    v.Groups = mutableListOf(byteArrayOf(0, 1))
    v.Named.value["n"] = -5
    v.When = Instant.ofEpochSecond(123456789, 123)
    v.P = Child().apply { ID = 3L; Text = "p"; Scores = byteArrayOf(6) }
    v.Dict["d"] = Child().apply { ID = 4L; Text = "z"; Scores = byteArrayOf(7) }
    val wire = ByteArray(v.size())
    check(v.marshal(wire, 0) == wire.size)
    val (end, copy) = Root.unmarshal(wire, 0)
    check(end == wire.size && copy.When == v.When)
    val again = ByteArray(copy.size())
    check(copy.marshal(again, 0) == again.size && wire.contentEquals(again))
    println("WIRE=" + wire.joinToString("") { "%02x".format(it.toInt() and 0xff) })
}
