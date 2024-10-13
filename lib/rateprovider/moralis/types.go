package moralis

type Token struct {
	TokenAddress string `json:"token_address"`
}

type Tokens struct {
	Tokens []Token `json:"tokens"`
}
