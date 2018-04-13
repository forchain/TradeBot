package helpers

import "github.com/forchain/cryptotrader/model"

func ReverseExchangeData(records []model.Record) []model.Record  {
	rLen:=len(records)
	if rLen<=0 {
		return records
	}
	var result []model.Record
	for _,x:=range records {
		result=append(result,model.Record(x))
	}
	for i:=0;i<rLen ;i++  {
		result[i].Close=records[rLen-i-1].Open
		result[i].Open=records[rLen-i-1].Close
		result[i].High=records[rLen-i-1].High
		result[i].Low=records[rLen-i-1].Low
		result[i].Vol=records[rLen-i-1].Vol
		result[i].Raw=records[rLen-i-1].Raw

	}
	return result
}