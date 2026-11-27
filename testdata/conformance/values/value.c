#include "model_benc.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
int main(void) {
 Root v={0}, copy={0};
 char s[3]={'A',0,'B'}; v.S.data=s; v.S.data_len=3;
 v.I=-7;v.U=300;v.Count=201;
 v.Fixed.item[0]=1;v.Fixed.item[1]=2;v.Fixed.item[2]=3;
 v.Numbers.item[0]=-2;v.Numbers.item[1]=5;
 uint8_t raw[2]={0,255};v.Raw.item=raw;v.Raw.item_count=2;
 Child child={0};child.ID=-9;child.Text.data="x";child.Text.data_len=1;uint8_t scores[2]={4,5};child.Scores.item=scores;child.Scores.item_count=2;v.Children.item=&child;v.Children.item_count=1;
 Child pointed={0};pointed.ID=3;pointed.Text.data="p";pointed.Text.data_len=1;uint8_t point_scores[1]={6};pointed.Scores.item=point_scores;pointed.Scores.item_count=1;v.P.item=&pointed;
 uint8_t group[2]={0,1};
 v.Groups.item=calloc(1,sizeof(*v.Groups.item));if(!v.Groups.item)return 4;
 v.Groups.item[0].item=group;v.Groups.item[0].item_count=2;v.Groups.item_count=1;
 v.Dict.item_keys=calloc(1,sizeof(*v.Dict.item_keys));if(!v.Dict.item_keys)return 6;
 v.Dict.item_keys[0].data="d";v.Dict.item_keys[0].data_len=1;
 Child mapped={0};mapped.ID=4;mapped.Text.data="z";mapped.Text.data_len=1;uint8_t map_scores[1]={7};mapped.Scores.item=map_scores;mapped.Scores.item_count=1;v.Dict.item_values=&mapped;v.Dict.item_count=1;
 v.Named.benc_value.item_keys=calloc(1,sizeof(*v.Named.benc_value.item_keys));if(!v.Named.benc_value.item_keys)return 5;
 v.Named.benc_value.item_keys[0].data="n";v.Named.benc_value.item_keys[0].data_len=1;
 int32_t val[1]={-5};v.Named.benc_value.item_values=val;v.Named.benc_value.item_count=1;
 v.When.unix_nano=123456789000000123;
 size_t size=Root_size(&v),off=0;uint8_t buf[256];if(size>sizeof(buf))return 1;
 if(Root_marshal(buf,size,&off,&v)!=BSTD_OK||off!=size)return 2;
 off=0;if(Root_unmarshal(buf,size,&off,&copy)!=BSTD_OK||off!=size)return 3;
 for(size_t i=0;i<size;i++){printf("%02x",buf[i]);} puts("");Root_free(&copy);free(v.Groups.item);free(v.Dict.item_keys);free(v.Named.benc_value.item_keys);return 0;
}
