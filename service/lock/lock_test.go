package lock

import (
	"testing"
	"log"
	"smartlockservice/crv"
)

func TestGetApplicationAuthor(t *testing.T){
	crvClinet:=&crv.CRVClient{
		Server:"http://127.0.0.1:8200",
		Token:"lockapi",
		AppID:"smartlockv3",
	}

	lockOperator:=&LockOperator{
		CRVClient:crvClinet,
	}

	author:=lockOperator.getAppAuthor("44")
	log.Println("author:",author)
	if author!="admin" {
		t.Fatalf("getAppAuthor failed")
	}
}

func TestDealAuthRec(t *testing.T){
	crvClinet:=&crv.CRVClient{
		Server:"http://127.0.0.1:8200",
		Token:"lockapi",
		AppID:"smartlockv3",
	}

	lockOperator:=&LockOperator{
		CRVClient:crvClinet,
	}

	lockOperator.DealAuthRec(&KCOperParm{
		OperType:KC_OPER_AUTHREC,
		ApplicationID:"44",
		KeyControllerID:"Box01",
		KeyID:"key01",
		Status:"0",
		Message:"test",
	})

}