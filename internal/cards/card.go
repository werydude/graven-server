package cards

type CardId string

type InstanceCard struct {
	CardId CardId `json:"CardID"`
	NodeId int    `json:"NodeID"`
}

type CardType interface {
	~struct {
		CardId CardId
		NodeId int
	}
}
