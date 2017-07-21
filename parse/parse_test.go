package parse

import (
	"testing"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

func TestParse(t *testing.T) {
	var res interface{}
	res = ParseIgnoreErr(bson.M{
		"_id":"aaaa",
		"hello":bson.M{"$gt":0},
		"hello2.g123":bson.M{"$in":[]interface{}{1,1}},
		"hello3":bson.M{"$eq":[]interface{}{"$regec",1}},
	},true)
	fmt.Println(res)

	res = ParseIgnoreErr([]bson.M{
		{
			"$match":bson.M{"_id":"21"},
		},{
			"$sort":bson.D{{"key","-1"},{"key2","-1"}},
		},{
			"$project":bson.M{
				"_id":1,
				"ss":bson.M{
					"$filter":bson.M{
						"input":"$as",
						"as":"$$as",
						"cond":bson.M{"$eq":[]interface{}{"$aa","ddd"}},
					},
				},
			},
		},{
			"$hhh":bson.RegEx{"kkk","dddd"},
		},
	},false)
	fmt.Println(res)


	fmt.Println(IsEqual(int64(3),int(3),false))
	fmt.Println(IsEqual(int64(1),int(3),false))
	fmt.Println(IsEqual("bs","bs",false))
	fmt.Println(IsEqual("a","b",false))
	fmt.Println(IsEqual(false,false,false))
	fmt.Println(IsEqual(false,true,false))
	fmt.Println(IsEqual([]int64{3,2,1},[]int64{1,2,3},false))
	fmt.Println(IsEqual([]int64{4,2,1},[]int64{1,2,3},false))
	fmt.Println(IsEqual([]string{"3","2","1"},[]string{"1","2","3"},false))
	fmt.Println(IsEqual([]string{"4","2","1"},[]string{"1","2","3"},false))
	fmt.Println(IsEqual([]bool{false,true},[]bool{true,false},false))
	fmt.Println(IsEqual([]bool{false},[]bool{true},false))



	res1,_ := Parse(bson.M{
		"_id":"string",
		"a":bson.M{"$in":[]string{"a","b"}},
		"b":bson.M{"$in":[]int{1,2}},
		"c":bson.M{"$in":[]bool{true}},
		"$sort":bson.D{{"hekko",-1},{"nima",1}},
		"$and":[]bson.M{
			{"$or":[]bson.M{
				{"a.g1":2,"b":true},
				{"c.u1":3,"d":false},
				{"a.g3":2,"b":false},
			}},
			{"$or":[]bson.M{
				{"c.u3":3,"d":false},
				{"a.g3":2,"b":false},
			}},
		},
	},true)
	res2,_ := Parse(bson.M{
		"_id":"string",
		"a":bson.M{"$in":[]string{"a","b"}},
		"b":bson.M{"$in":[]int{1,2}},
		"c":bson.M{"$in":[]interface{}{true}},
		"$sort":bson.D{{"nima",-1},{"hekko",1}},
		"$and":[]bson.M{
			{"$or":[]bson.M{
				{"c.u5":3,"d":false},
				{"a.g100":2,"b":false},
			}},
		},
	},true)

	fmt.Println(IsEqual(res1,res2,true))


	res1,_ = Parse([]bson.M{
		{
			"$match":bson.M{
				"_id":"string",
				"a":bson.M{"$in":[]string{"a","b"}},
				"b":bson.M{"$in":[]int{1,2}},
				"c":bson.M{"$in":[]bool{true}},
				"$sort":bson.D{{"hekko",-1},{"nima",1}},
				"$and":[]bson.M{
					{"$or":[]bson.M{
						{"a.g1":2,"b":true},
						{"c.u1":3,"d":false},
						{"a.g3":2,"b":false},
					}},
					{"$or":[]bson.M{
						{"c.u3":3,"d":false},
						{"a.g3":2,"b":false},
					}},
				},
			},
		},{
			"$sort":bson.D{{"a",1},{"b",2}},
		},{
			"$limit":100,
		},{
			"$project":bson.M{
				"_id":1,
				"uid":"$uid2",
			},
		},{
			"$group":bson.M{
				"_id":"$uid",
				"uids":bson.M{"$addToSet":"$uid"},
			},
		},
	},true)
	fmt.Println(res1)
}
