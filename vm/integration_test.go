package vm

import (
	"testing"

	"fmt"

	"os"

	"io/ioutil"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/vm/database"
	"github.com/iost-official/go-iost/vm/host"
	"github.com/iost-official/go-iost/vm/native"
	. "github.com/smartystreets/goconvey/convey"
)

var testID = []string{
	"IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE",
	"IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "8dJ9YKovJ5E7hkebAQaScaG1BA8snRUHPUbVcArcTVq6",
	"IOST7ZNDWeh8pHytAZdpgvp7vMpjZSSe5mUUKxDm6AXPsbdgDMAYhs", "7CnwT7BXkEFAVx6QZqC7gkDhQwbvC3d2CkMZvXHZdDMN",
	"IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd", "Htarc5Sp4trjqY4WrTLtZ85CF6qx87v7CRwtV4RRGnbF",
	"IOST7GmPn8xC1RESMRS6a62RmBcCdwKbKvk2ZpxZpcXdUPoJdapnnh", "Bk8bAyG4VLBcrsoRErPuQGhwCy4C1VxfKE4jjX9oLhv",
	"IOST7ZGQL4k85v4wAxWngmow7JcX4QFQ4mtLNjgvRrEnEuCkGSBEHN", "546aCDG9igGgZqVZeybajaorP5ZeF9ghLu2oLncXk3d6",
	"IOST59uMX3Y4ab5dcq8p1wMXodANccJcj2efbcDThtkw6egvcni5L9", "DXNYRwG7dRFkbWzMNEbKfBhuS8Yn51x9J6XuTdNwB11M",
	"IOST8mFxe4kq9XciDtURFZJ8E76B8UssBgRVFA5gZN9HF5kLUVZ1BB", "AG8uECmAwFis8uxTdWqcgGD9tGDwoP6CxqhkhpuCdSeC",
	"IOST7uqa5UQPVT9ongTv6KmqDYKdVYSx4DV2reui4nuC5mm5vBt3D9", "GJt5WSSv5WZi1axd3qkb1vLEfxCEgKGupcXf45b5tERU",
	"IOST6wYBsLZmzJv22FmHAYBBsTzmV1p1mtHQwkTK9AjCH9Tg5Le4i4", "7U3uwEeGc2TF3Xde2oT66eTx1Uw15qRqYuTnMd3NNjai",
}

var systemContract = native.SystemABI()

func ininit(t *testing.T) (*database.Visitor, db.MVCCDB) {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		t.Fatal(err)
	}

	//mvccdb := replaceDB(t)

	vi := database.NewVisitor(0, mvccdb)
	vi.SetTokenBalance("iost", testID[0], 1000000)
	vi.SetContract(systemContract)
	vi.Commit()

	return vi, mvccdb
}

func closeMVCCDB(m db.MVCCDB) {
	m.Close()
	os.RemoveAll("mvcc")
}

func MakeTx(act *tx.Action) (*tx.Tx, error) {
	trx := tx.NewTx([]*tx.Action{act}, nil, 100000, 100, 10000000, 0)

	ac, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	if err != nil {
		return nil, err
	}
	trx, err = tx.SignTx(trx, ac.ID, []*account.KeyPair{ac})
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func MakeTxWithAuth(act *tx.Action, ac *account.KeyPair) (*tx.Tx, error) {
	trx := tx.NewTx([]*tx.Action{act}, nil, 100000, 100, 10000000, 0)
	trx, err := tx.SignTx(trx, ac.ID, []*account.KeyPair{ac})
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func TestIntergration_Transfer(t *testing.T) {
	// TODO: new test

	/*  t.Skip() */
	// ilog.Stop()
	// e, vi, mvcc := ininit(t)
	// defer closeMVCCDB(mvcc)

	// act := tx.NewAction("system.iost", "Transfer", fmt.Sprintf(`["%v","%v","%v"]`, testID[0], testID[2], "0.000001"))

	// trx := tx.NewTx([]*tx.Action{act}, nil, 10000, 100, 10000000, 0)

	// ac, err := account.NewKeyPair(common.Base58Decode(testID[1]), crypto.Secp256k1)
	// if err != nil {
	// t.Fatal(err)
	// }
	// trx, err = tx.SignTx(trx, ac.ID, []*account.KeyPair{ac})
	// if err != nil {
	// t.Fatal(err)
	// }

	// Convey("trasfer success case", t, func() {
	// r, err := e.Exec(trx, time.Second)
	// if r.Status.Code != 0 {
	// t.Fatal(r)
	// }
	// So(err, ShouldBeNil)
	// So(vi.TokenBalance("iost", testID[0]), ShouldEqual, int64(999597))
	// So(vi.TokenBalance("iost", testID[2]), ShouldEqual, int64(100))
	// })

	// act2 := tx.NewAction("system.iost", "Transfer", fmt.Sprintf(`["%v","%v",%v]`, testID[0], testID[2], "999896"))
	// trx2 := tx.NewTx([]*tx.Action{act2}, nil, 10000, 100, 10000000, 0)
	// trx2, err = tx.SignTx(trx2, ac.ID, []*account.KeyPair{ac})
	// if err != nil {
	// t.Fatal(err)
	// }

	// Convey("trasfer balance not enough case", t, func() {
	// r, err := e.Exec(trx2, time.Second)
	// if r.Status.Code != 4 {
	// t.Fatal(r)
	// }
	// So(err, ShouldBeNil)
	// So(vi.TokenBalance("iost", testID[0]), ShouldEqual, int64(999586))
	// So(vi.TokenBalance("iost", testID[2]), ShouldEqual, int64(100))
	/* }) */
}

func jsHelloWorld() *contract.Contract {
	jshw := contract.Contract{
		ID: "ContractjsHelloWorld",
		Code: `
class Contract {
 init() {

 }
 hello() {
  return "world";
 }
}

module.exports = Contract;
`,
		Info: &contract.Info{
			Lang:    "javascript",
			Version: "1.0.0",
			Abi: []*contract.ABI{
				{
					Name: "hello",
					Args: []string{},
				}, {
					Name: "constructor",
					Args: []string{},
				},
			},
		},
	}
	return &jshw
}

/* func TestEngine_InitSetCode(t *testing.T) { */
// t.Skip("dep")

// mvccdb, err := db.NewMVCCDB("mvcc")
// if err != nil {
// t.Fatal(err)
// }

// defer closeMVCCDB(mvccdb)

// vi := database.NewVisitor(0, mvccdb)
// vi.SetTokenBalance("iost", testID[0], 1000000)
// vi.SetContract(systemContract)
// vi.Commit()

// bh := &block.BlockHead{
// ParentHash: []byte("abc"),
// Number:     0,
// Witness:    "witness",
// Time:       123456,
// }

// e := newEngine(bh, vi)

// //e.SetUp("js_path", jsPath)
// e.SetUp("log_level", "debug")
// e.SetUp("log_enable", "")

// jshw := jsHelloWorld()

// act := tx.NewAction("system.iost", "InitSetCode", fmt.Sprintf(`["test.iost", "%v"]`, jshw.B64Encode()))

// trx, err := MakeTx(act)
// if err != nil {
// t.Fatal(err)
// }

// r, err := e.Exec(trx, time.Second)
// if err != nil {
// t.Fatal(err)
// }
// if r.Status.Code != tx.Success {
// t.Fatal(r)
// }
// ilog.Debugf(fmt.Sprintln("balance of sender :", vi.TokenBalance("iost", testID[0])))

// act2 := tx.NewAction("test.iost", "hello", `[]`)

// trx2, err := MakeTx(act2)
// if err != nil {
// t.Fatal(err)
// }

// r, err = e.Exec(trx2, time.Second)
// if err != nil {
// t.Fatal(err)
// }
// if r.Status.Code != tx.Success {
// t.Fatal(r)
// }
// ilog.Debugf(fmt.Sprintln("balance of sender :", vi.TokenBalance("iost", testID[0])))
/* } */

func TestIntergration_CallJSCode(t *testing.T) {
	// TODO: new test

	/*  t.Skip("dep") */

	// ilog.Stop()
	// e, vi, mvcc := ininit(t)
	// defer closeMVCCDB(mvcc)

	// jshw := jsHelloWorld()
	// jsc := jsCallHelloWorld()

	// vi.SetContract(jshw)
	// vi.SetContract(jsc)

	// act := tx.NewAction("Contractcall_hello_world", "call_hello", fmt.Sprintf(`[]`))

	// trx, err := MakeTx(act)
	// if err != nil {
	// t.Fatal(err)
	// }

	// r, err := e.Exec(trx, time.Second)
	// if err != nil {
	// t.Fatal(err)
	// }
	// if r.Status.Code != 0 {
	// t.Fatal(r.Status.Message)
	// }
	// if vi.TokenBalance("iost", testID[0]) != int64(1000000) { // todo something wrong here!
	// t.Fatal(vi.TokenBalance("iost", testID[0]))
	/* } */
}

func jsCallHelloWorld() *contract.Contract {
	return &contract.Contract{
		ID: "Contractcall_hello_world",
		Code: `
class Contract {
 init() {

 }
 call_hello() {
  return BlockChain.call("ContractjsHelloWorld", "hello", "[]")
 }
}

module.exports = Contract;
`,
		Info: &contract.Info{
			Lang:    "javascript",
			Version: "1.0.0",
			Abi: []*contract.ABI{
				{
					Name: "call_hello",
					Args: []string{},
				},
			},
		},
	}
}

func TestIntergration_CallJSCodeWithReceipt(t *testing.T) {
	// TODO: new test

	/*  t.Skip("dep") */

	// ilog.Stop()
	// e, vi, mvcc := ininit(t)
	// defer closeMVCCDB(mvcc)

	// jshw := jsHelloWorld()
	// jsc := jsCallHelloWorldWithReceipt()

	// vi.SetContract(jshw)
	// vi.SetContract(jsc)

	// act := tx.NewAction("Contractcall_hello_world", "call_hello", fmt.Sprintf(`[]`))

	// trx, err := MakeTx(act)
	// if err != nil {
	// t.Fatal(err)
	// }

	// r, err := e.Exec(trx, time.Second)
	// if err != nil {
	// t.Fatal(err)
	// }
	// if r.Status.Code != 0 {
	// t.Fatal(r.Status.Message)
	// }
	// if vi.TokenBalance("iost", testID[0]) != int64(999999) {
	// t.Fatal(vi.TokenBalance("iost", testID[0]))
	/* } */
}

func jsCallHelloWorldWithReceipt() *contract.Contract {
	return &contract.Contract{
		ID: "Contractcall_hello_world",
		Code: `
class Contract {
 init() {

 }
 call_hello() {
  return BlockChain.callWithReceipt("ContractjsHelloWorld", "hello", "[]")
 }
}

module.exports = Contract;
`,
		Info: &contract.Info{
			Lang:    "javascript",
			Version: "1.0.0",
			Abi: []*contract.ABI{
				{
					Name: "call_hello",
					Args: []string{},
				},
			},
		},
	}
}

func TestIntergration_Payment_Success(t *testing.T) {
	// TODO: new test

	/*  t.Skip("dep") */

	// jshw := jsHelloWorld()

	// //ilog.Debugf("init %v", jshw.Info.Abis[0].GetLimit())

	// e, vi, mvcc := ininit(t)
	// defer closeMVCCDB(mvcc)
	// vi.SetContract(jshw)

	// vi.SetTokenBalance("iost", "CGjsHelloWorld", 1000000)

	// act := tx.NewAction("ContractjsHelloWorld", "hello", fmt.Sprintf(`[]`))

	// trx, err := MakeTx(act)
	// if err != nil {
	// t.Fatal(err)
	// }

	// r, err := e.Exec(trx, time.Second)
	// if err != nil {
	// t.Fatal(err)
	// }
	// if r.Status.Code != 0 {
	// t.Fatal(r.Status.Message)
	// }
	// if vi.TokenBalance("iost", testID[0]) != int64(1000000) {
	// t.Fatal(vi.TokenBalance("iost", testID[0]))
	// }
	// if vi.TokenBalance("iost", "CGjsHelloWorld") != int64(1000000) { // todo something wrong here
	// t.Fatal(vi.TokenBalance("iost", "CGjsHelloWorld"))
	// }

}

func TestIntergration_Payment_Failed(t *testing.T) {
	// TODO: new test

	/*  t.Skip("dep") */
	// jshw := jsHelloWorld()

	// e, vi, mvcc := ininit(t)
	// defer closeMVCCDB(mvcc)
	// vi.SetContract(jshw)

	// vi.SetTokenBalance("iost", "CGjsHelloWorld", 1000000)
	// vi.Commit()

	// act := tx.NewAction("ContractjsHelloWorld", "hello", fmt.Sprintf(`[]`))

	// trx, err := MakeTx(act)
	// if err != nil {
	// t.Fatal(err)
	// }

	// r, err := e.Exec(trx, time.Second)
	// ilog.Debugf("success: %v, %v", r, err)
	// ilog.Debugf("balance of sender : %v", vi.TokenBalance("iost", testID[0]))
	// ilog.Debugf("balance of contract : %v", vi.TokenBalance("iost", "CGjsHelloWorld"))

}

type fataler interface {
	Fatal(args ...interface{})
}

type JSTester struct {
	t      fataler
	vi     *database.Visitor
	mvccdb db.MVCCDB

	cname string
	c     *contract.Contract
}

func NewJSTester(t fataler) *JSTester {
	mvccdb, err := db.NewMVCCDB("mvcc")
	if err != nil {
		panic(err)
	}

	//mvccdb := replaceDB(t)

	vi := database.NewVisitor(0, mvccdb)
	vi.SetTokenBalance("iost", testID[0], 1000000*1e8)
	vi.SetContract(systemContract)
	vi.Commit()

	return &JSTester{
		t:      t,
		vi:     vi,
		mvccdb: mvccdb,
	}
}

func (j *JSTester) ReadDB(key string) (value interface{}) {
	return database.MustUnmarshal(j.vi.Get(j.cname + "-" + key))
}

func (j *JSTester) ReadMap(key, field string) (value interface{}) {
	return database.MustUnmarshal(j.vi.MGet(j.cname+"-"+key, field))
}

func (j *JSTester) FlushDB(t *testing.T, keys []string) {
	for _, k := range keys {
		t.Logf("%s: %v", k, j.ReadDB(k))
	}
}

func (j *JSTester) NewBlock(bh *block.BlockHead) {
}

func (j *JSTester) SetJS(code string) {
	j.c = &contract.Contract{
		ID:   "jsContract",
		Code: code,
		Info: &contract.Info{
			Lang:    "javascript",
			Version: "1.0.0",
			Abi: []*contract.ABI{
				{
					Name: "constructor",
					Args: []string{},
				},
			},
		},
	}
}

func (j *JSTester) DoSet() *tx.TxReceipt {
	return nil
	/*  act := tx.NewAction("system.iost", "SetCode", fmt.Sprintf(`["%v"]`, j.c.B64Encode())) */

	// trx, err := MakeTx(act)
	// if err != nil {
	// j.t.Fatal(err)
	// }
	// r, err := j.e.Exec(trx, time.Second)
	// if err != nil {
	// j.t.Fatal(err)
	// }
	// j.cname = "Contract" + common.Base58Encode(trx.Hash())

	/* return r */
}

func (j *JSTester) SetAPI(name string, argType ...string) {

	j.c.Info.Abi = append(j.c.Info.Abi, &contract.ABI{
		Name: name,
		Args: argType,
	})

}

func (j *JSTester) TestJS(main, args string) *tx.TxReceipt {

	return nil
	// act2 := tx.NewAction(j.cname, main, args)

	// trx2, err := MakeTx(act2)
	// if err != nil {
	// j.t.Fatal(err)
	// }

	// r, err := j.e.Exec(trx2, time.Second)
	// if err != nil {
	// j.t.Fatal(err)
	// }
	/* return r */
}

func (j *JSTester) TestJSWithAuth(abi, args, seckey string) *tx.TxReceipt {
	return nil

	/* act2 := tx.NewAction(j.cname, abi, args) */

	// ac, err := account.NewKeyPair(common.Base58Decode(seckey), crypto.Secp256k1)
	// if err != nil {
	// panic(err)
	// }

	// trx2, err := MakeTxWithAuth(act2, ac)
	// if err != nil {
	// j.t.Fatal(err)
	// }

	// r, err := j.e.Exec(trx2, time.Second)
	// if err != nil {
	// j.t.Fatal(err)
	// }
	/* return r */
}

func (j *JSTester) Clear() {
	j.mvccdb.Close()
	os.RemoveAll("mvcc")
}

func TestJSAPI_Database(t *testing.T) {
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	this.aa = new Int64(100);
	}
	main() {
		this.aa = new Int64(45);
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of publisher :", js.vi.TokenBalance("iost", testID[0]))
	t.Log("balance of receiver :", js.vi.TokenBalance("iost", testID[2]))
	t.Log("value of this.aa :", js.ReadDB("aa"))
}

func TestJSAPI_Transfer(t *testing.T) {
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	}
	main() {
		BlockChain.transfer("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "100", "")
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.TokenBalance("iost", testID[0]))
	t.Log("balance of receiver :", js.vi.TokenBalance("iost", testID[2]))
}

func TestJSAPI_Transfer_Failed(t *testing.T) {
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	}
	main() {
		BlockChain.transfer("IOST54ETA3q5eC8jAoEpfRAToiuc6Fjs5oqEahzghWkmEYs9S9CMKd", "IOST558jUpQvBD7F3WTKpnDAWg6HwKrfFiZ7AqhPFf4QSrmjdmBGeY", "100", "")
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.TokenBalance("iost", testID[0]))
	t.Log("balance of receiver :", js.vi.TokenBalance("iost", testID[2]))
}

func TestJSAPI_Transfer_WrongFormat1(t *testing.T) {
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	}
	main() {
		var ret = BlockChain.transfer("a", "b", 1, "");
		if (ret !== 0) {
			throw new Error("ret = ", ret);
		}
	}
}

module.exports = Contract;
`)
	js.SetAPI("main")
	js.DoSet()

	r := js.TestJS("main", fmt.Sprintf(`[]`))
	//todo wrong receipt
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.TokenBalance("iost", testID[0]))
	t.Log("balance of receiver :", js.vi.TokenBalance("iost", testID[2]))
}

func TestJSAPI_Deposit(t *testing.T) {
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	}
	deposit() {
		return BlockChain.deposit("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "100", "")
	}
	withdraw() {
		return BlockChain.withdraw("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn", "99", "")
	}
}

module.exports = Contract;
`)
	js.SetAPI("deposit")
	js.SetAPI("withdraw")
	js.DoSet()

	r := js.TestJS("deposit", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.TokenBalance("iost", testID[0]))
	if 100*1e8 != js.vi.TokenBalance("iost", host.ContractAccountPrefix+js.cname) {
		t.Fatal(js.vi.TokenBalance("iost", host.ContractAccountPrefix+js.cname))
		t.Fatalf("balance of contract " + js.cname + "should be 100.")
	}

	r = js.TestJS("withdraw", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
	t.Log("balance of sender :", js.vi.TokenBalance("iost", testID[0]))
	if 1*1e8 != js.vi.TokenBalance("iost", host.ContractAccountPrefix+js.cname) {
		t.Fatalf("balance of contract " + js.cname + "should be 1.")
	}
}

func TestJSAPI_Info(t *testing.T) {
	ilog.Stop()
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	}
	blockInfo() {
		var info = BlockChain.blockInfo()
		var obj = JSON.parse(info)
		console.log(obj["parent_hash"])
		console.log(obj.number)
		return obj["parent_hash"]
	}
	txInfo() {
		var info = BlockChain.txInfo()
		var obj = JSON.parse(info)
		console.log(obj["hash"])
		return obj["hash"]
	}
}

module.exports = Contract;
`)
	js.SetAPI("blockInfo")
	js.SetAPI("txInfo")
	js.DoSet()

	r := js.TestJS("blockInfo", fmt.Sprintf(`[]`))
	if r.Status.Code != 0 {
		t.Fatal(r.Status.Message)
	}

	r = js.TestJS("txInfo", fmt.Sprintf(`[]`))
	if r.Status.Code != 0 {
		t.Fatal(r.Status.Message)
	}
}

func TestJSRequireAuth(t *testing.T) {
	t.Skip("dep")

	js := NewJSTester(t)
	defer js.Clear()

	js.SetJS(`
class Contract {
	init() {
	}
	requireAuth() {
		var ok = BlockChain.requireAuth("haha")
		_native_log(JSON.stringify(ok))
		ok = BlockChain.requireAuth("IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn")
		_native_log(JSON.stringify(ok))
		return ok
	}
}

module.exports = Contract;
`)
	js.SetAPI("requireAuth")
	js.DoSet()

	r := js.TestJS("requireAuth", fmt.Sprintf(`[]`))
	t.Log("receipt is ", r)
}

func TestJS_Database(t *testing.T) {
	t.Skip("dep")
	js := NewJSTester(t)
	defer js.Clear()

	lc, err := ioutil.ReadFile("test_data/database.js")
	if err != nil {
		t.Fatal(err)
	}
	js.SetJS(string(lc))
	js.SetAPI("read")
	js.SetAPI("change")
	js.DoSet()
	//t.Log("========= constructor")
	Convey("test of js database", t, func() {
		So(js.ReadDB("num").(string), ShouldEqual, "9")
		So(js.ReadDB("string").(string), ShouldEqual, "hello")
		So(js.ReadDB("bool").(string), ShouldEqual, "true")
		So(js.ReadDB("array").(string), ShouldEqual, "[1,2,3]")
		So(js.ReadDB("obj").(string), ShouldEqual, `{"foo":"bar"}`)
	})
	r := js.TestJS("read", `[]`)
	if r.Status.Code != 0 {
		t.Fatal(r.Status.Message)
	}
	//js.TestJS("change", `[]`)
	////t.Log("========= change")
	////t.Log("array is ", js.ReadDB("array"))
	////t.Log("object is ", js.ReadDB("object"))
	////t.Log("arrayobj is ", js.ReadDB("arrayobj"))
	////t.Log("objobj is ", js.ReadDB("objobj"))
	////t.Log("keyobj is", js.ReadDB("key"))
}

/*
func TestJS_LuckyBet(t *testing.T) {
	ilog.Stop()

	js := NewJSTester(t)
	defer js.Clear()
	lc, err := ReadFile("test_data/lucky_bet.js")
	if err != nil {
		t.Fatal(err)
	}
	js.vi.SetTokenBalance("iost",testID[0], 100000000000000)
	js.SetJS(string(lc))
	js.SetAPI("clearUserValue")
	js.SetAPI("bet", "string", "number", "number", "number")
	js.SetAPI("getReward")
	r := js.DoSet()
	if r.Status.Code != 0 {
		t.Fatal(r.Status.Message)
	}

	// here put the first bet
	r = js.TestJS("bet", fmt.Sprintf(`["%v",0, 200000000, 1]`, testID[0]))
	Convey("after 1 bet", t, func() {
		So(r.Status.Message, ShouldEqual, "")
		So(js.ReadDB("user_number"), ShouldEqual, "1")
		So(js.ReadDB("total_coins"), ShouldEqual, "200000000")
		So(js.ReadMap("table", "0"), ShouldEqual, `[{"account":"IOST4wQ6HPkSrtDRYi2TGkyMJZAB3em26fx79qR3UJC7fcxpL87wTn","coins":200000000,"nonce":1}]`)
	})

	for i := 1; i < 100; i++ { // at i = 2, should get reward
		r = js.TestJS("bet", fmt.Sprintf(`["%v",%v,%v,%v]`, testID[0], i%10, (i%4+1)*100000000, i))
		if r.Status.Code != 0 {
			t.Fatal(r.Status.Message)
		}
		if r.GasUsage < 1000 {
			t.Fatal(r.GasUsage)
		}
	}

	Convey("after 100 bet", t, func() {
		So(r.Status.Message, ShouldEqual, "")
		So(js.ReadDB("user_number"), ShouldEqual, "0")
		So(js.ReadDB("total_coins"), ShouldEqual, "0")
		So(js.ReadDB("round"), ShouldEqual, "2")
		So(js.ReadDB("result1"), ShouldContainSubstring, `{"number":200,"user_number":100,"k_number":10,"total_coins":{"number":"23845000000"},`)
		t.Log(js.vi.TokenBalance("iost","CA"+js.cname), js.cname)
	})
}
*/
