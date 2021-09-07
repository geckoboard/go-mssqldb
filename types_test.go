package mssql

import (
	"reflect"
	"testing"
	"time"
)

var fakeType uint8 = 0xFF

func TestMakeGoLangScanType(t *testing.T) {
	if (reflect.TypeOf(int64(0)) != makeGoLangScanType(typeInfo{TypeId: typeInt8})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(float64(0)) != makeGoLangScanType(typeInfo{TypeId: typeFlt4})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(float64(0)) != makeGoLangScanType(typeInfo{TypeId: typeFlt8})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf("") != makeGoLangScanType(typeInfo{TypeId: typeVarChar})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(time.Time{}) != makeGoLangScanType(typeInfo{TypeId: typeDateTime})) {
		t.Errorf("invalid type returned for typeDateTime")
	}
	if (reflect.TypeOf(time.Time{}) != makeGoLangScanType(typeInfo{TypeId: typeDateTim4})) {
		t.Errorf("invalid type returned for typeDateTim4")
	}
	if (reflect.TypeOf(int64(0)) != makeGoLangScanType(typeInfo{TypeId: typeInt1})) {
		t.Errorf("invalid type returned for typeInt1")
	}
	if (reflect.TypeOf(int64(0)) != makeGoLangScanType(typeInfo{TypeId: typeInt2})) {
		t.Errorf("invalid type returned for typeInt2")
	}
	if (reflect.TypeOf(int64(0)) != makeGoLangScanType(typeInfo{TypeId: typeInt4})) {
		t.Errorf("invalid type returned for typeInt4")
	}
	if (reflect.TypeOf(int64(0)) != makeGoLangScanType(typeInfo{TypeId: typeIntN, Size: 4})) {
		t.Errorf("invalid type returned for typeIntN")
	}
	if (reflect.TypeOf([]byte{}) != makeGoLangScanType(typeInfo{TypeId: typeMoney, Size: 8})) {
		t.Errorf("invalid type returned for typeIntN")
	}
	if (reflect.TypeOf(nil) != makeGoLangScanType(typeInfo{TypeId: typeUdt})) {
		t.Errorf("invalid type returned for user defined type")
	}
	if (reflect.TypeOf(nil) != makeGoLangScanType(typeInfo{TypeId: fakeType})) {
		t.Errorf("invalid type returned for unhandled type")
	}
}

func TestMakeGoLangTypeName(t *testing.T) {
	t.Run("check multiple types", func(t *testing.T) {
		tests := []struct {
			typeName   string
			typeString string
			typeID     uint8
		}{
			{"typeDateTime", "DATETIME", typeDateTime},
			{"typeDateTim4", "SMALLDATETIME", typeDateTim4},
			{"typeBigBinary", "BINARY", typeBigBinary},
			{"unhandled type", "UNHANDLED", fakeType},
			//TODO: Add other supported types
		}

		for _, tt := range tests {
			if makeGoLangTypeName(typeInfo{TypeId: tt.typeID}) != tt.typeString {
				t.Errorf("invalid type name returned for %s", tt.typeName)
			}
		}
	})

	t.Run("returns user defined type name", func(t *testing.T) {
		want := "GEOGRAPHY"
		ti := typeInfo{
			UdtInfo: udtInfo{TypeName: "geography"},
			TypeId:  typeUdt,
			Size:    0,
		}

		if got := makeGoLangTypeName(ti); got != want {
			t.Errorf("unexpected type wanted %q but got %q", want, got)
		}
	})
}

func TestMakeGoLangTypeLength(t *testing.T) {
	tests := []struct {
		typeName   string
		typeVarLen bool
		typeLen    int64
		typeID     uint8
		size       int
	}{
		{"typeDateTime", false, 0, typeDateTime, 0},
		{"typeDateTim4", false, 0, typeDateTim4, 0},
		{"typeBigVarChar", true, 2147483645, typeBigVarChar, 0xffff},
		{"typeBigVarChar", true, 10, typeBigVarChar, 10},
		{"typeBigBinary", true, 30, typeBigBinary, 30},
		{"userDefinedType", false, 0, typeUdt, 0},
		{"unhandledType", false, 0, fakeType, 0},
		//TODO: Add other supported types
	}

	for _, tt := range tests {
		n, v := makeGoLangTypeLength(typeInfo{TypeId: tt.typeID, Size: tt.size})
		if v != tt.typeVarLen {
			t.Errorf("invalid type length variability returned for %s", tt.typeName)
		}
		if n != tt.typeLen {
			t.Errorf("invalid type length returned for %s", tt.typeName)
		}
	}
}

func TestMakeGoLangTypePrecisionScale(t *testing.T) {
	tests := []struct {
		typeName   string
		typeID     uint8
		typeVarLen bool
		typePrec   int64
		typeScale  int64
	}{
		{"typeDateTime", typeDateTime, false, 0, 0},
		{"typeDateTim4", typeDateTim4, false, 0, 0},
		{"typeBigBinary", typeBigBinary, false, 0, 0},
		{"userDefinedType", typeUdt, false, 0, 0},
		{"userDefinedType", fakeType, false, 0, 0},
		//TODO: Add other supported types
	}

	for _, tt := range tests {
		prec, scale, varLen := makeGoLangTypePrecisionScale(typeInfo{TypeId: tt.typeID})
		if varLen != tt.typeVarLen {
			t.Errorf("invalid type length variability returned for %s", tt.typeName)
		}
		if prec != tt.typePrec || scale != tt.typeScale {
			t.Errorf("invalid type precision and/or scale returned for %s", tt.typeName)
		}
	}
}

func TestMakeDecl(t *testing.T) {
	t.Run("check mulitple types", func(t *testing.T) {
		tests := []struct {
			typeName string
			Size     int
			typeID   uint8
		}{
			{"varchar(max)", 0xffff, typeVarChar},
			{"varchar(8000)", 8000, typeVarChar},
			{"varchar(4001)", 4001, typeVarChar},
			{"nvarchar(max)", 0xffff, typeNVarChar},
			{"nvarchar(4000)", 8000, typeNVarChar},
			{"nvarchar(2001)", 4002, typeNVarChar},
			{"varbinary(max)", 0xffff, typeBigVarBin},
			{"varbinary(8000)", 8000, typeBigVarBin},
			{"varbinary(4001)", 4001, typeBigVarBin},
			{"unhandled", 0, fakeType},
		}

		for _, tt := range tests {
			s := makeDecl(typeInfo{TypeId: tt.typeID, Size: tt.Size})
			if s != tt.typeName {
				t.Errorf("invalid type translation for %s", tt.typeName)
			}
		}
	})

	t.Run("returns user defined type name", func(t *testing.T) {
		want := "custom type"
		ti := typeInfo{
			UdtInfo: udtInfo{TypeName: want},
			TypeId:  typeUdt,
			Size:    0,
		}

		if got := makeDecl(ti); got != want {
			t.Errorf("unexpected type wanted %q but got %q", want, got)
		}
	})
}
