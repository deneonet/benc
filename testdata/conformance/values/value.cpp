#include "model_benc.hpp"
#include <cstdio>
using namespace models;
int main(){
 Root v{};v.S=std::string("A\0B",3);v.I=-7;v.U=300;v.Count=201;v.Fixed={1,2,3};v.Numbers={-2,5};v.Raw={std::byte{0},std::byte{255}};
 Child ch{};ch.ID=-9;ch.Text="x";ch.Scores={std::byte{4},std::byte{5}};v.Children={ch};v.Groups={{std::byte{0},std::byte{1}}};v.Named["n"]=-5;
 Child pointed{};pointed.ID=3;pointed.Text="p";pointed.Scores={std::byte{6}};v.P=pointed;
 Child mapped{};mapped.ID=4;mapped.Text="z";mapped.Scores={std::byte{7}};v.Dict["d"]=mapped;
 v.When=bstd::time_point{std::chrono::nanoseconds{123456789000000123LL}};
 std::vector<std::byte> b(v.Size());if(v.Marshal(b,0)!=b.size())return 1;
 Root copy{};auto r=copy.Unmarshal(b,0);if(!std::holds_alternative<std::size_t>(r)||std::get<std::size_t>(r)!=b.size())return 2;
 for(auto x:b){std::printf("%02x",std::to_integer<unsigned>(x));}std::puts("");return 0;
}
