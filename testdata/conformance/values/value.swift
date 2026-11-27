import Foundation
import Bstd

@main enum BencValue {
  static func main() throws {
    var v = Root()
    v.S = "A\0B"; v.I = -7; v.U = 300; v.Count = 201
    v.Fixed = [1, 2, 3]; v.Numbers = [-2, 5]; v.Raw = [0, 255]
    var ch = Child(); ch.ID = -9; ch.Text = "x"; ch.Scores = [4, 5]; v.Children = [ch]
    v.Groups = [[0, 1]]; v.Named.value["n"] = -5
    v.When = 123456789000000123
    var pointed = Child(); pointed.ID = 3; pointed.Text = "p"; pointed.Scores = [6]; v.P = pointed
    var mapped = Child(); mapped.ID = 4; mapped.Text = "z"; mapped.Scores = [7]; v.Dict["d"] = mapped
    var wire = [UInt8](repeating: 0, count: v.size()); var n = 0
    try v.marshal(at: &n, into: &wire); precondition(n == wire.count)
    n = 0; let copy = try Root.unmarshal(at: &n, from: wire)
    precondition(n == wire.count && copy.When == v.When)
    var again = [UInt8](repeating: 0, count: copy.size()); n = 0
    try copy.marshal(at: &n, into: &again); precondition(n == again.count && again == wire)
    print("WIRE=" + wire.map { String(format: "%02x", $0) }.joined())
  }
}
