package main

import "testing"
import "reflect"

func TestPosFindingString(t *testing.T) {
	variates := stringSlice{"one", "two", "three"}
	if x := variates.pos("two"); x != 1 {
		t.Errorf("Expected 'two' to be at postion 1 but go %s", x)
	}
}
func TestPostNotFindingString(t *testing.T) {
	variates := stringSlice{"one", "two", "three"}
	if x := variates.pos("nothing"); x != -1 {
		t.Errorf("Expected 'two' to be at postion -1 but go %s", x)
	}
}

func TestPrepareData(t *testing.T) {
	expected := map[string]string{
		"one":   "1",
		"two":   "2",
		"three": "3",
	}
	fields := []string{"one", "two", "three"}
	variates := stringSlice{"1", "2", "3"}
	x := prepareData(fields, variates)
	if !reflect.DeepEqual(x, expected) {
		t.Errorf("Expected %s but got %s", expected, x)
	}
}

// TODO: figure out how to test the import
func TestRowToHash(t *testing.T) {
	// input := `"TOA5","raingauge","CR1000","21942","CR1000.Std.16","CPU:NOAHIV_2_3_0.CR1","51123","Table1"
	// "TIMESTAMP","RECORD","GageMinV","ActTemp","ActDepth","ReportPCP","collMinVolt(1)","Cycles(1)","WetTime(1)","DryTime(1)","MissedTime(1)","collMinVolt(2)","Cycles(2)","WetTime(2)","DryTime(2)","MissedTime(2)","collMinVolt(3)","Cycles(3)","WetTime(3)","DryTime(3)","MissedTime(3)","OPDCounts","operatingMode15","blockedSec","scan10","keepAliveSec","ActDepthRA"
	// "TS","RN","","","","","","","","","","","","","","","","","","","","","","","","",""
	// "","","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp","Smp"
	// "2015-08-08 15:15:00",136541,12.02,25.92,0.2951884,0,-1.414,0,0,0,0,-0.259,0,0,0,0,-1.181,0,0,0,0,0,0,0,900,0,0.2949845
	// "2015-08-08 15:30:00",136542,12.06,26.16,0.2943513,0,-1.401,0,0,0,0,-0.263,0,0,0,0,-1.176,0,0,0,0,1,0,0,900,0,0.2949069
	// "2015-08-08 15:45:00",136543,12.05,26.59,0.2927501,0,-1.397,0,0,0,0,-0.263,0,0,0,0,-1.172,0,0,0,0,2,0,0,900,0,0.2929058`

}
