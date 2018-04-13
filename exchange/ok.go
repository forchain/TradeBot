package exchange

type OkEx struct {
	Exchange

}
func (p *OkEx) Init(){

}
func (p *OkEx) Exit(){

}
func (p *OkEx) GetExchange()(*Exchange){
	return &p.Exchange
}
func (p *OkEx)Execute(cmd OptRecord) {

}