package util

import (
	"fmt"
	"github.com/system-pclub/GCatch/GCatch/config"
	"github.com/system-pclub/GCatch/GCatch/tools/go/ssa"
	"go/token"
	"go/types"
	"reflect"
	"strings"
)

var mapStruct2Pointer map[* types.Struct] * types.Pointer

func GetStructPointerMapping()  {
	mapStruct2Pointer = make(map[* types.Struct] * types.Pointer)
	for _, T := range config.Prog.RuntimeTypes() {
		if pPointer, ok := T.(* types.Pointer); ok {
			if pNamed, ok := pPointer.Elem().(* types.Named); ok {
				if pStruct, ok := pNamed.Underlying().(* types.Struct); ok {
					mapStruct2Pointer[pStruct] = pPointer
				}
			}
		}
	}
	//fmt.Println("mapping size", len(mapStruct2Pointer))
}

/*
func GetInterfaceMethods() {
	for _, pkg := range config.Prog.AllPackages() {
		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				//fmt.Println(fn.String())
				if fn.Name() == "measure" {
					fn.WriteTo(os.Stdout)
				}
			}
		}
	}


	for _, T := range config.Prog.RuntimeTypes() {
		mset := config.Prog.MethodSets.MethodSet(T)
		for i, n := 0, mset.Len(); i < n; i++ {
			//visit.function(visit.prog.MethodValue(mset.At(i)))
			fmt.Println(config.Prog.MethodValue(mset.At(i)))
		}
	}

}
*/

/*
func GetFieldTypes(t  types.Type, m map[types.Type] bool) {
	if pPointer, ok := t.(* types.Pointer); ok {
		GetFieldTypes(pPointer.Elem(), m)
	} else if pNamed, ok := t.(* types.Named); ok {
		GetFieldTypes(pNamed.Underlying(), m)
	} else if pStruct, ok := t.(* types.Struct); ok {
		m[pStruct] = true
		for i := 0; i < pStruct.NumFields(); i ++ {
			GetFieldTypes(pStruct.Field(i).Type(), m)
		}
	} else if pInterface, ok := t.(* types.Interface); ok {
		m[pInterface] = true
		for i := 0; i < pInterface.NumEmbeddeds(); i ++ {
			GetFieldTypes(pInterface.EmbeddedType(i), m)
		}

	} else if pMap, ok := t.(*types.Map); ok {
		m[pMap] = true
	} else if pBasic, ok := t.(*types.Basic); ok {
		m[pBasic] = true
	} else {
		fmt.Println("not handle", reflect.TypeOf(t), t)
		panic("in GetFieldTypes()")
	}
}

func PrintTypes(m map[types.Type] bool) {
	for t, _ := range m {
		fmt.Println(reflect.TypeOf(t), t)
	}
}
*/

func GetTypeMethods(t types.Type, m map[string] bool, mVisited map[types.Type] bool)  {

	if _, ok := mVisited[t]; ok {
		return
	}

	mVisited[t] = true

	if pPointer, ok := t.(* types.Pointer); ok {
		GetTypeMethods(pPointer.Elem(), m, mVisited)
	} else if pNamed, ok := t.(* types.Named); ok {
		GetTypeMethods(pNamed.Underlying(), m, mVisited)
	} else if pStruct, ok := t.(* types.Struct); ok {
		if pPointer, ok := mapStruct2Pointer[pStruct]; ok {
			mset := config.Prog.MethodSets.MethodSet(pPointer)
			if mset != nil {
				for i := 0; i < mset.Len(); i++ {
					if pFunc, ok := mset.At(i).Obj().(*types.Func); ok {
						m[pFunc.FullName()] = true
					}
				}
			}
		}
		for i := 0; i < pStruct.NumFields(); i ++ {
			GetTypeMethods(pStruct.Field(i).Type(), m, mVisited)
		}
	} else if pInterface, ok := t.(* types.Interface); ok {
		for i := 0; i < pInterface.NumMethods(); i ++ {
			m[pInterface.Method(i).FullName()] = true
		}
		for i := 0; i < pInterface.NumEmbeddeds(); i ++ {
			GetTypeMethods(pInterface.EmbeddedType(i), m, mVisited)
		}

	} else if _, ok := t.(*types.Map); ok {
	} else if _, ok := t.(*types.Basic); ok {
	} else if _, ok := t.(*types.Slice); ok {
	} else if _, ok := t.(*types.Chan); ok {
	} else if _, ok := t.(*types.Array); ok {
	} else if _, ok := t.(*types.Signature); ok {
	}else {
		fmt.Println("not handle", reflect.TypeOf(t), t)
		panic("in GetTypeMethods()")
	}
}

func DecoupleTypeMethods(m map[string] bool) map[string] map[string] bool {
	mapResult := make(map[string] map[string] bool)
	for strFunName, _ := range m {
		var strStructName string
		var strFName string
		if strings.LastIndex(strFunName, ".") < 0 {
			strStructName = ""
			strFName = strFunName
		} else {
			strStructName = strFunName[:strings.LastIndex(strFunName, ".")]
			strFName = strFunName[strings.LastIndex(strFunName, ".") + 1:]
		}

		if _, ok := mapResult[strStructName]; !ok {
			mapResult[strStructName] = make(map[string] bool)
		}

		mapResult[strStructName][strFName] = true
	}

	return mapResult
}

func PrintTypeMethods(m map[string] map[string] bool) {
	for sPackage, mapMethods := range m {
		fmt.Println(sPackage)
		for sFuncName, _ := range mapMethods {
			fmt.Println(sFuncName)
		}
	}
}

func GetBaseType(v ssa.Value) types.Type {
	if i, ok := v.(ssa.Instruction); ok {
		if f, ok := i.(* ssa.FieldAddr); ok {
			return GetBaseType(f.X)
		} else if u, ok := i.(* ssa.UnOp); ok {
			switch u.Op {
			case token.MUL:
				return GetBaseType(u.X)
			}
		}

	}

	return v.Type()
}

type SyncStruct struct {
	Pkg string
	TypeName string
	Type string
	ContainSync bool
	Field map[string]string
	Potential bool
}

func ListAllSyncStruct(prog *ssa.Program) []*SyncStruct {
	allSyncStruct := *new([]*SyncStruct)

	allPotentialStruct := *new([]*SyncStruct)
	for _, pkg := range prog.AllPackages() { //loop all packages
		if pkg == nil {
			continue
		}


		for memName, _ := range pkg.Members { //loop through all members; the member may be a func or a type; if it is type, loop through all its methods

			//check if this member is a type
			memAsType := pkg.Type(memName)
			if memAsType != nil {
				//this member is a type

				newStructPtr := findStructPotential(memAsType)
				if newStructPtr == nil {
					continue
				}

				if newStructPtr.Potential == true {
					allPotentialStruct = append(allPotentialStruct, newStructPtr)
				}
			}
		}
	}



	//now all potential structs are listed
	//at this point, each struct in all_potential_struct has struct.potential == true, meaning it may be moved to all_sync_struct
	//append structs that ContainSync = true, to all_sync_struct

	for _, syncStruct := range allPotentialStruct {
		if (*syncStruct).ContainSync {
			allSyncStruct = append(allSyncStruct, syncStruct)
			(*syncStruct).Potential = false
		}
	}


	for {
		flagBreak := true

	potential:
		for _, structPotential := range allPotentialStruct {
			if structPotential.Potential == false { //this struct has already been moved to all_sync_struct
				continue
			}
			if structPotential.TypeName == "node" && structPotential.Pkg == "v2store" {
				print()
			}

			for _, fieldType := range structPotential.Field {
				for _, structSync := range allSyncStruct {
					if structPotential.TypeName == "node" && structPotential.Pkg == "v2store" && structSync.TypeName == "store" && structSync.Pkg == "v2store" {
						print()
					}
					if strings.Contains(fieldType, structSync.Type) { // This struct contains a field, which is in all_sync_struct
						allSyncStruct = append(allSyncStruct, structPotential)
						structPotential.Potential = false
						flagBreak = false // don't break the infinite loop since there is still at least one struct moving from all_potential_struct to all_sync_struct
						continue potential
					}
				}
			}
		}

		if flagBreak == true {
			break
		}

	}

	return allSyncStruct
}

func findStructPotential(memAsType *ssa.Type) (result *SyncStruct) {
	result = nil
	defer func() {
		if r := recover(); r != nil {
			result = nil
		}
	}()
	structType := memAsType.String()
	structContainsync := false
	structPotential := false

	fieldsStr := memAsType.Object().Type().Underlying().String()
	if !strings.HasPrefix(fieldsStr,"struct{") || fieldsStr == "struct{}" {
		return nil
	}
	fieldsStr = strings.Replace(fieldsStr,"struct{","",1)
	fieldsStr = fieldsStr[:len(fieldsStr) - 1] //delete the last char, which is "}"
	fields := strings.Split(fieldsStr,"; ")
	if len(fields) == 0  {
		return nil
	} else {
		str := strings.ReplaceAll(fields[0]," ","")
		if len(str) == 0 {
			return nil
		}
	}

	structField := make(map[string]string)
	for index,field := range fields {
		fieldElement := strings.Split(field," ")
		var fieldName, fieldType string
		if len(fieldElement) == 1 { //anonymous field
			fieldName = string(index)
			fieldType = fieldElement[0]
			if fieldElement[0] == "chan" && len(fieldElement) > 1 {
				fieldType = "chan " + fieldElement[1]
			}
		} else {
			fieldName = fieldElement[0]
			fieldType = fieldElement[1]
			if fieldElement[1] == "chan" && len(fieldElement) > 2 {
				fieldType = "chan " + fieldElement[2]
			}
		}
		structField[fieldName] = fieldType
		if isTypeSync(fieldType) {
			structContainsync = true
			structPotential = true
		}
		if strings.Contains(fieldType,"/") {
			structPotential = true
		}
	}
	newStructPtr := &SyncStruct{
		Pkg:         "",
		TypeName:    memAsType.Name(),
		Type:        structType,
		ContainSync: structContainsync,
		Field:       structField,
		Potential:   structPotential,
	}
	if memAsType.Package() != nil {
		if memAsType.Package().Pkg != nil {
			newStructPtr.Pkg = memAsType.Package().Pkg.Name()
		}
	}
	result = newStructPtr
	return
}

func isTypeSync(str string) bool {
	str = strings.ReplaceAll(str,"*","")
	str = strings.ReplaceAll(str,"[]","")
	if strings.HasPrefix(str,"sync.") || strings.Contains(str,"<-") || strings.HasPrefix(str,"chan ") {
		return true
	} else if strings.Contains(str,"map[") {
		if strings.Contains(str,"sync.") || strings.Contains(str,"<-") || strings.Contains(str,"chan ") {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}