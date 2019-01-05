package triasapp

import (
	"fmt"
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
	TypeTransaction = "trans"
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
	postValue := url.Values{
		"tx": {string(tx)},
	}
	resp, err := http.PostForm("http://127.0.0.1:9981/deliver_tx", postValue)

	if err != nil{
		return types.ErrUnknownRequest
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "success" {
		return types.OK
	}else {
		return types.ErrBaseInsufficientFunds
	}

}

// 检查数据
// 如果包括合约相关的，不关心是合约的创建还是执行，都转发到合约处
// 如果包括帐户交易相关的，直接转到交易服务处
func (app *TriasCodeApplication) CheckTx(tx []byte) types.Result {
	log.Println("[triascode]: CheckTx -----------")
	jsonStr := string(tx)
	log.Println("[triascode] the tx is :", jsonStr)

	if strings.Contains(jsonStr, TypeContract) {
		log.Println("[triascode] the type is contract, the json is ", TypeContract, jsonStr)
		//cm := ttypes.CodeMessage{}
		//error := json.Unmarshal(tx, &cm)
		return app.executeContract(jsonStr)
	}

	if strings.Contains(string(tx), TypeTransaction) {
		log.Println("[triascode] the type is contract, the json is ", TypeTransaction, jsonStr)
		return app.checkUtxoTransaction(tx)

	}
	return types.OK
}

func (app *TriasCodeApplication) Commit() types.Result {
	hash := app.state.Hash()
	fmt.Println("---------commit-------")
	return types.NewResultOK(hash, "")
}

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

// callback by http post
// @param data: message
func (app *TriasCodeApplication) executeContract(jsonStr string) (types.Result) {
	url := "http://127.0.0.1:8088/executeContract"
	//url := "http://13.250.34.43:8088/executeContract"
	payload := strings.NewReader(jsonStr)
	log.Println("param :" , payload)
	req, err1 := http.NewRequest("POST", url, payload)
	if err1 != nil {
		log.Println(err1)
		return types.ErrUnmashallJson
	}
	if req == nil {
		return types.ErrUnmashallJson
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	//req.Header.Add("Postman-Token", "d101fb95-a0f1-407d-b217-6284407825d2")
	log.Println("begain to req ")
	res, err := http.DefaultClient.Do(req)
	if err != nil  {
		log.Println(err)
		return types.ErrUnmashallJson
	}
	if res == nil {
		return types.ErrUnmashallJson
	}

	defer res.Body.Close()
	body, error1 := ioutil.ReadAll(res.Body)
	log.Println("body ------",body)
	if error1 != nil {
		log.Println(error1)
		return types.ErrUnmashallJson
	} else {
		log.Println(res)
		log.Println("get body from fb: ", string(body))
		return types.OK
	}

}

// callback by http post
// @param data: tx
func (app *TriasCodeApplication) checkUtxoTransaction(tx []byte) (types.Result) {
	postValue := url.Values{
		"tx": {string(tx)},
	}

	resp, err := http.PostForm("http://127.0.0.1:9981/check_tx", postValue)
	if err != nil{
		return types.ErrUnmashallJson
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "success" {
		return types.OK
	}else {
		return types.ErrBaseInvalidSignature
	}
}
