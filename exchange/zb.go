package exchange

type ZBEx struct {
	Exchange

}
func (p *ZBEx) Init(){

}
func (p *ZBEx) Exit(){

}
func (p *ZBEx) GetExchange()(*Exchange){
	return &p.Exchange
}
func (p *ZBEx)Execute(cmd OptRecord) {

}