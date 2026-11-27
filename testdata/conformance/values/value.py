from model_benc import Root, Child

v = Root()
v.S = "A\0B"
v.I = -7
v.U = 300
v.Count = 201
v.Fixed = bytes([1, 2, 3])
v.Numbers = [-2, 5]
v.Raw = bytes([0, 255])
v.Children = [Child(ID=-9, Text="x", Scores=bytes([4, 5]))]
v.Groups = [bytes([0, 1])]
v.Named["n"] = -5
v.When = 123456789000000123
v.P = Child(ID=3, Text="p", Scores=bytes([6]))
v.Dict["d"] = Child(ID=4, Text="z", Scores=bytes([7]))

wire = bytearray(v.size())
assert v.marshal(0, wire) == len(wire)
end, copy = Root.unmarshal(0, wire)
assert end == len(wire) and copy == v and copy.When == v.When
print(wire.hex())
