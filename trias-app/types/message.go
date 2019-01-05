package types

type  CodeMessage struct {
	Address 			string		`json:"address"`
	CheckMD5   			string		`json:"checkMD5"`
	Command    			string		`json:"command"`
	ContractName    	string		`json:"contractName"`
	ContractType    	string		`json:"contractType"`
	ContractVersion 	string		`json:"contractVersion"`
	VmVersion   		string		`json:"vmVersion"`
	Sequence			string		`json:"sequence"`
	User    			string		`json:"user"`
	Signature    		string		`json:"signature"`
	Operation    		string		`json:"operation"`
	Timestamp			int			`json:"timestamp"`
}

type TransactionMessage struct {
	FromAddress 		string		`json:"from_address"`
	ToAddress 			string		`json:"to_address"`
	Amount 				string		`json:"amount"`
}
