package lockhub

import (
	"testing"
	"log"
)

func TestGetLockNoHexStr(t *testing.T){
	lockNo:="110"
	lockNoHexStr:=GetLockNoHexStr(lockNo)
	log.Println("lockNoHexStr:",lockNoHexStr)
	lockNoNew:=GetLockNoDecStr(lockNoHexStr)
	log.Println("lockNoNew:",lockNoNew)
	if lockNo!=lockNoNew {
		t.Fatalf("lockNo!=lockNoNew")
	}
}