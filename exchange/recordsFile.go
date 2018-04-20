package exchange

import (
	"os"
	"time"
	"strings"
	"strconv"
)

type RecordsFileInfo struct {
	File os.FileInfo
	Time time.Time
}
func (p *RecordsFileInfo)Init()error{

	var err error
	var t int64
	strTemp:=strings.Split(p.File.Name(),"_")
	if len(strTemp)==2 {

		//
		strTemp=strings.Split(strTemp[1],".")
		if len(strTemp)==2 {
			t,err=strconv.ParseInt(strTemp[0],10,64)
			if err != nil {
				return err
			}
			p.Time=time.Unix(t,0)
		}


	}

	return nil
}
type RecordsFileInfoSort []RecordsFileInfo
func (p RecordsFileInfoSort) Len() int {
	return len(p)
}
func (p RecordsFileInfoSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p RecordsFileInfoSort) Less(i, j int) bool {
	return p[i].Time.Before(p[j].Time)
}
