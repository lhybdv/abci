package triasapp

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
)

const (
	TypeContract    = "contract"
	TypeTransaction = "Vout"
)

type TriasCodeApplication struct {
	types.BaseApplication

	state merkle.Tree
}

func NewTriasCodeApplication() *TriasCodeApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &TriasCodeApplication{state: state}
}

func (app *TriasCodeApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *TriasCodeApplication) DeliverTx(tx []byte) types.Result {
	jsonStr := string(tx)
	log.Println("[triascode]-----------deliver tx :", jsonStr)

	//parts := strings.Split(string(tx), "=")
	//if len(parts) == 2 {
	//	log.Println("parts =2")
	//	app.state.Set([]byte(parts[0]), []byte(parts[1]))
	//} else {
	//	log.Println("parts =1", tx)
	//	app.state.Set(tx, tx)
	//}

	// TODO json格式化代码contains来判断请求头
	if strings.Contains(jsonStr, TypeContract) {
		log.Println("[triascode] the type is TypeContract, the json is ", TypeContract, jsonStr)
		result := app.deliverContract(jsonStr)
		if result.Code == types.OK.Code {
			app.state.Set(tx, tx)
		}
		return result
	}
	if strings.Contains(jsonStr, TypeTransaction) {
		log.Println("[triascode] the type is TypeTransaction, the json is ", TypeTransaction, jsonStr)
		result := app.deliverUtxoTransaction(tx)
		if result.Code == types.OK.Code {
			app.state.Set(tx, tx)
		}
		return result
	}
	app.state.Set(tx, tx)
	return types.OK
}

// 检查数据
// 如果包括合约相关的，不关心是合约的创建还是执行，都转发到合约处
// 如果包括帐户交易相关的，直接转到交易服务处
func (app *TriasCodeApplication) CheckTx(tx []byte) types.Result {
	log.Println("[triascode]: CheckTx -----------")
	jsonStr := string(tx)
	log.Println("[triascode] 29 the tx is :", jsonStr)
	if strings.Contains(jsonStr, TypeContract) {
		// TODO json格式化代码contains来判断请求头
		log.Println("[triascode] the type is contract----, the json is ", TypeContract, jsonStr)
		//cm := ttypes.CodeMessage{}
		//error := json.Unmarshal(tx, &cm)
		checkRe := app.checkContract(tx)
		log.Println("check contract result:", checkRe)
		return checkRe
	} else if strings.Contains(jsonStr, TypeTransaction) {
		log.Println("[triascode] the type is utxo trans -----, the json is ", jsonStr)
		checkRe := app.checkUtxoTransaction(tx)
		log.Println("check utxo result:", checkRe)
		return checkRe
	}

	return types.OK
}

func (app *TriasCodeApplication) Commit() types.Result {
	log.Println("[triascode]---------commit-------")
	//app.state.Save()
	hash := app.state.Hash()
	log.Println("[triascode]commit hash is :", hash)
	return types.NewResultOK(hash, "")
	//resp, err := http.Get("http://127.0.0.1:9981/commit_tx")
	//if err != nil {
	//	return types.NewError(6,"")
	//}
	//defer resp.Body.Close()
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil{
	//	return types.NewError(2,"")
	//}
	//body_byte, _ := hex.DecodeString(string(body))
	//return types.NewResultOK(body_byte, "")
}

// TODO 查询merkle tree的交易问题修复
func (app *TriasCodeApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	if reqQuery.Prove {
		value, proof, exists := app.state.Proof(reqQuery.Data)
		resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		resQuery.Proof = proof
		if exists {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	} else {
		index, value, exists := app.state.Get(reqQuery.Data)
		resQuery.Index = int64(index)
		resQuery.Value = value
		if exists {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}

func (app *TriasCodeApplication) checkContract(tx []byte) types.Result {
	return types.OK
}

// callback by http post
// @param data: message
func (app *TriasCodeApplication) deliverContract(jsonStr string) types.Result {
	url := "http://127.0.0.1:8088/executeContract"
	payload := strings.NewReader(jsonStr)
	log.Println("the deliver contract url is :", url)
	log.Println("param :", payload)
	req, err1 := http.NewRequest("POST", url, payload)
	if err1 != nil {
		log.Println(err1)
		return types.ErrContractExecute
	}
	if req == nil {
		return types.ErrContractExecute
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	//req.Header.Add("Postman-Token", "d101fb95-a0f1-407d-b217-6284407825d2")
	log.Println("begin to request  contract ")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Do request err: ", err)
		return types.ErrContractExecute
	}
	if res == nil {
		return types.ErrContractExecute
	}

	defer res.Body.Close()
	body, error1 := ioutil.ReadAll(res.Body)
	log.Println("request from contract ,the body ------", body)
	if error1 != nil {
		log.Println("read contract body err: ", error1)
		return types.ErrContractExecute
	} else {
		strBody := string(body)
		log.Println("get body from body of contract : ", strBody)
		if strings.Contains(strBody, "success") {
			return types.OK
		}
		return types.ErrContractExecute
	}

}

// callback by http post
// @param data: tx
func (app *TriasCodeApplication) checkUtxoTransaction(tx []byte) types.Result {
	postValue := url.Values{
		"tx": {string(tx)},
	}
	resp, err := http.PostForm("http://127.0.0.1:9981/check_tx", postValue)
	if err != nil {
		log.Println("[Triascode] error:", err)
		return types.ErrCheckTrans
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("utxo check_tx result :", string(body))
	if string(body) == "success" {
		log.Println("check from utxo sucess")
		return types.OK
	} else {
		log.Println("check from utxo failed")
		return types.ErrCheckTrans
	}
}

// callback by http post
// @param data: tx
func (app *TriasCodeApplication) deliverUtxoTransaction(tx []byte) types.Result {
	postValue := url.Values{
		"tx": {string(tx)},
	}

	resp, err := http.PostForm("http://127.0.0.1:9981/deliver_tx", postValue)
	//resp, err := http.PostForm("http://192.168.1.11:9981/deliver_tx", postValue)
	if err != nil {
		return types.ErrDeliverTrans
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("deliver utxo result: ", body)
	if string(body) == "success" {
		log.Println("deliver sucess")
		return types.OK
	} else {
		log.Println("deliver failed")
		return types.ErrDeliverTrans
	}
}
