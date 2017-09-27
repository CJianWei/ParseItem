package parse

import (
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"errors"
	"sort"
	"math"
	"sync"
)

var Replace = map[string]string{
	"u": "UID",
	"g": "GID",
}

type NT int

const (
	B      NT = iota //bool
	I                //int
	S                //string
	BM               //bson.M
	BS               // bool s
	IS               // int s
	SS               // string s
	BMS              //bson.M s
	ES               // multiple type in array
	UNKNOW           //unknow type
)

//判断对象类型
func TypeOfV(v interface{}) (interface{}, NT) {
	if v == nil{
		return nil,UNKNOW
	}
	k := reflect.TypeOf(v)
	switch k.Kind() {
	case reflect.Bool:
		return true, B
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return 0, I
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 0, I
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return 0, I
	case reflect.String:
		res,_ := v.(string)
		if strings.HasPrefix(res,"$"){
			return RegexCommon(res, ".", Replace),S
		}
		return "string",S
	case reflect.Struct:
		if res, ok := v.(bson.DocElem); ok {
			return bson.M{RegexCommon(res.Name, ".", Replace): 1}, BM
		}
		if res, ok := v.(bson.RegEx); ok {
			return bson.M{"key": res.Options}, BM
		}
	case reflect.Map:
		if res,ok:=v.(bson.M);ok{
			return res,BM
		}
		var bm bson.M
		bys,_ := json.Marshal(v)
		err := json.Unmarshal(bys,&bm)
		if err != nil{
			return nil, UNKNOW
		}
		return bm, BM
	case reflect.Slice:
		if res, ok := v.(bson.D); ok {
			var bms = []bson.M{}
			for i:=0;i<len(res);i++{
				bms = append(bms,bson.M{RegexCommon(res[i].Name, ".", Replace): 1})
			}
			return bms, BMS
		}
		var vs = TranslateVs(v)
		if len(vs) <= 0{
			return nil,UNKNOW
		}
		if IsCommonAryType(vs){
			_,ns := TypeOfV(vs[0])
			switch ns {
			case B:
				return nil,BS
			case I:
				return nil,IS
			case S:
				var ress = []string{}
				for _,sin_s := range vs{
					str,_ := sin_s.(string)
					ress = append(ress,str)
				}
				return ress,SS
			case BM:
				var bms = []bson.M{}
				for _,sin_m := range vs{
					str,_ := sin_m.(bson.M)
					bms = append(bms,str)
				}
				return bms,BMS
			}
		}
		return vs,ES
	default:

	}
	return nil, UNKNOW
}

//替换掉某个个关键字 例如 u123 g123 替换为 UID,GID
func RegexCommon(str string, split string, replace map[string]string) string {
	arys := strings.Split(str, split)
	res := []string{}
	for _, sin_str := range arys {
		res = append(res, ReplacePre(sin_str, replace))
	}
	return strings.Join(res, split)
}

//替换掉指定的元素
func ReplacePre(str string, replace map[string]string) string {
	for k, v := range replace {
		if strings.HasPrefix(str, k) && Isdigitle(strings.Replace(str, k, "", 1)) {
			return v
		}
	}
	return str
}

//判断是否数字
func Isdigitle(str string) bool {
	_, err := strconv.Atoi(str)
	if err != nil {
		return false
	} else {
		return true
	}
}

//判断interface 数组中的元素是否保留一致类型
func IsCommonAryType(vs []interface{}) bool {
	record := map[NT]int{}
	for _,v :=range vs{
		_,nt := TypeOfV(v)
		record[nt] ++
	}
	if len(record) > 1{
		return false
	}else{
		return true
	}
}

func ParseIgnoreErr(v interface{},DuplicateRemove bool)interface{}{
	if v==nil{
		return "NIL"
	}
	res,err := Parse(v,DuplicateRemove)
	if err != nil{
		return "UNKNOW"
	}
	return res
}

func Parse(v interface{},DuplicateRemove bool) (interface{},error) {
	if v == nil{
		return nil,errors.New("NIL")
	}
	if res,ok:= v.(bson.M);ok{
		return ParseV(res,DuplicateRemove)
	}
	if res,ok:= v.([]bson.M);ok{
		return ParseV(res,DuplicateRemove)
	}
	res,nt := TypeOfV(v)
	if nt == BM || nt == BMS{
		return ParseV(res,DuplicateRemove)
	}
	return nil,errors.New("the current paramge can only deal with baon.M []bson.M util.Map []util.Map")
}

func ParseV(v interface{},DuplicateRemove bool) (interface{},error) {
	if v == nil{
		return nil,errors.New("nil of v")
	}
	if bm,ok:= v.(bson.M);ok{
		var tmp = bson.M{}
		for k,value :=range bm{
			k_tmp := RegexCommon(k, ".", Replace)
			nd,ns := TypeOfV(value)
			switch ns {
			case B:
				tmp[k_tmp]="bool"
			case I:
				tmp[k_tmp]="int"
			case BS:
				tmp[k_tmp]="[bool]"
			case IS:
				tmp[k_tmp]="[int]"
			case S:
				tmp[k_tmp]= nd
			case SS:
				tmp[k_tmp]= ResultOfArySS(nd)
			case ES:
				tmp[k_tmp]= ResultOfAryES(nd)
			case BM:
				if bm_tmp,err := ParseV(nd,false);err == nil{
					tmp[k_tmp]=bm_tmp
				}else{
					tmp[k_tmp]="UNKNOW"
				}
			case BMS:
				var bl = false
				if k_tmp == "$and" || k_tmp == "$or"{
					bl = true
				}
				if bms_tmp,err := ParseV(nd,bl);err==nil{
					tmp[k_tmp]=bms_tmp
				}else{
					tmp[k_tmp]="UNKNOW"
				}
			default:
				tmp[k_tmp]="UNKNOW"
			}
		}
		return tmp,nil
	}
	if bms,ok := v.([]bson.M);ok{
		var idx = map[int]bson.M{}
		var index = 0
		var tmps = []bson.M{}
		for _,sin_m :=range bms{
			tmp_sin,err := ParseV(sin_m,false)
			if err != nil{
				continue
			}
			tmp_sin_bm,_ := tmp_sin.(bson.M)
			idx[index] = tmp_sin_bm
			index++
			tmps = append(tmps,tmp_sin_bm)
		}

		if DuplicateRemove{
			for index,bm :=range tmps{
				for index2,bm2:=range idx{
					if index != index2{
						if IsEqual(bm,bm2,true){
							delete(idx,index)
						}
					}
				}
			}
		}

		//此处去重
		tmps = []bson.M{}
		for i:=0;i< len(bms);i++{
			if bm,ok:=idx[i];ok{
				tmps = append(tmps,bm)
			}
		}
		//for _,bm :=range idx{
		//	tmps = append(tmps,bm)
		//}
		return tmps,nil
	}
	return nil,errors.New("unknow type")
}

func ResultOfArySS(v interface{}) interface{} {
	strs,ok := v.([]string)
	if ok{
		var spec = []string{}
		for _,str := range strs{
			if strings.HasPrefix(str,"$"){
				spec = append(spec,str)
			}
		}
		if len(spec) > 0{
			sort.Strings(spec)
			spec = append(spec,"string")
			return spec
		}
		return "[string]"
	}
	return "UNKNOW"

}

func ResultOfAryES(v interface{}) interface{} {
	if reflect.TypeOf(v).Kind() != reflect.Slice{
		return "UNKNOW"
	}
	var vs = TranslateVs(v)
	var spec = []string{}
	for _,sin_v :=range vs{
		nd,nt := TypeOfV(sin_v)
		if nt == S {
			str,_ := nd.(string)
			if strings.HasPrefix(str,"$"){
				spec = append(spec,str)
			}
		}
	}
	if len(spec)> 0{
		sort.Strings(spec)
		spec = append(spec,"not string")
		return spec
	}
	return "ES"

}

func IsEqual(v1 interface{},v2 interface{},translateSort bool) bool {
	if v1 == nil || v2 == nil{
		return false
	}
	_,v1t := TypeOfV(v1)
	_,v2t :=  TypeOfV(v2)
	if !(v1t == v2t && v1t != UNKNOW){
		return false
	}
	_,nt := TypeOfV(v1)
	switch nt {
	case I:
		return FloatOwn(v1) - FloatOwn(v2) == 0
	case S:
		v1s,_:= v1.(string)
		v2s,_:= v2.(string)
		return strings.Contains(v1s,v2s) && strings.Contains(v2s,v1s)
	case B:
		v1b,_:= v1.(bool)
		v2b,_:= v2.(bool)
		return v1b == v2b
	case IS:
		sortF := func(v interface{}) []float64 {
			var float64_ary = []float64{}
			for _,sin_v :=range TranslateVs(v){
				float64_ary = append(float64_ary,FloatOwn(sin_v))
			}
			sort.Float64s(float64_ary)
			return float64_ary
		}
		var vs1 = sortF(v1)
		var vs2 = sortF(v2)

		if len(vs1) != len(vs2){
			return false
		}
		for i := 0;i<len(vs1);i++{
			if IsEqual(vs1[i],vs2[i],translateSort) == false{
				return false
			}
		}
		return true
	case SS:
		sortS := func(v interface{})[]string {
			var strs = []string{}
			for _,sin_v :=range TranslateVs(v){
				str,_ := sin_v.(string)
				strs = append(strs,str)
			}
			sort.Strings(strs)
			return strs
		}
		var vs1 = sortS(v1)
		var vs2 = sortS(v2)
		if len(vs1) == len(vs2){
			for i := 0;i<len(vs1);i++{
				if IsEqual(vs1[i],vs2[i],translateSort) ==false{
					return false
				}
			}
			return true
		}
	case BS:
		sortB := func(v interface{})map[bool]int {
			b_m := map[bool]int{}
			for _,sin_v :=range TranslateVs(v){
				b,_ := sin_v.(bool)
				b_m[b] ++
			}
			return b_m
		}
		var vs1 = sortB(v1)
		var vs2 = sortB(v2)
		for b,_ :=range vs1{
			if vs2[b] <= 0{
				return false
			}
		}
		return true
	case BM:
		var bm1,_ = v1.(bson.M)
		var bm2,_ = v2.(bson.M)
		for k,v_ :=range bm1{
			if v2_,ok:=bm2[k];ok{
				if IsEqual(v_,v2_,translateSort) ==false{
					return false
				}
			}else{
				return false
			}
		}
		return true
	case BMS:
		getBs := func(v interface{}) []bson.M{
			bms := []bson.M{}
			vs := TranslateVs(v)
			for _,sin_v :=range vs{
				bm,_ := sin_v.(bson.M)
				bms = append(bms,bm)
			}
			return bms
		}
		var vs1 = getBs(v1)
		var vs2 = getBs(v2)

		if translateSort == false{
			if len(vs1)!= len(vs2){
				return false
			}
			for i:=0;i< len(vs1);i++{
				if IsEqual(vs1[i],vs2[i],true) == false{
					return false
				}
			}
			return true
		}else{
			dump := func(vs []bson.M) []bson.M {
				var idx_index = map[int]int{}
				for i:=0;i<len(vs)-1;i++{
					for j:= i+1;j<len(vs);j++{
						if IsEqual(vs[i],vs[j],true) ==true{
							idx_index[i]++
							break
						}
					}
				}
				var res = []bson.M{}
				for i:=0;i<len(vs);i++{
					if idx_index[i] > 0{
						continue
					}
					res = append(res,vs[i])
				}
				return res
			}
			vs1 = dump(vs1)
			vs2 = dump(vs2)
			if len(vs1)!= len(vs2){
				return false
			}
			var count int = 0
			for i:=0;i< len(vs1);i++{
				for j:= 0;j< len(vs2);j++{
					if IsEqual(vs1[i],vs2[j],true) == true{
						count ++
					}
				}
			}
			if count == len(vs1){
				return true
			}else{
				return false
			}
		}
	default:
	}
	return false
}

func TranslateVs( v interface{}) []interface{} {
	vals := reflect.ValueOf(v)
	var vs = []interface{}{}
	for i := 0; i < vals.Len(); i++ {
		vs = append(vs, vals.Index(i).Interface())
	}
	return vs

}

func FloatOwn(v interface{}) (float64) {
	if v == nil{
		return float64(math.MinInt64)
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int:
		return float64(v.(int))
	case reflect.Int8:
		return float64(v.(int8))
	case reflect.Int16:
		return float64(v.(int16))
	case reflect.Int32:
		return float64(v.(int32))
	case reflect.Int64:
		return float64(v.(int64))
	case reflect.Uint:
		return float64(v.(uint))
	case reflect.Uint8:
		return float64(v.(uint8))
	case reflect.Uint16:
		return float64(v.(uint16))
	case reflect.Uint32:
		return float64(v.(uint32))
	case reflect.Uint64:
		return float64(v.(uint64))
	case reflect.Float32:
		return float64(v.(float32))
	case reflect.Float64:
		return float64(v.(float64))
	default:
		return float64(math.MinInt64)
	}
}

func ToString(v interface{}) string {
	bys,_ := json.Marshal(v)
	return string(bys)
}

var R = map[string]interface{}{}
var lck = sync.RWMutex{}

func GetString(v interface{})string{
	lck.Lock()
	defer lck.Unlock()
	for str,v_tmp :=range R{
		if IsEqual(v,v_tmp,true) == true{
			return str
		}
	}
	s := ToString(v)
	R[s]=v
	return s
}

