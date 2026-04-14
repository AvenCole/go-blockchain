package gui

type DashboardData struct {
	Height         int    `json:"height"`
	LatestHash     string `json:"latestHash"`
	MerkleRoot     string `json:"merkleRoot"`
	Difficulty     int    `json:"difficulty"`
	Nonce          int    `json:"nonce"`
	PendingTxCount int    `json:"pendingTxCount"`
	WalletCount    int    `json:"walletCount"`
	DataDir        string `json:"dataDir"`
	NetworkMode    string `json:"networkMode"`
}

type WalletView struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type BlockView struct {
	Height           int               `json:"height"`
	Hash             string            `json:"hash"`
	PrevHash         string            `json:"prevHash"`
	MerkleRoot       string            `json:"merkleRoot"`
	Difficulty       int               `json:"difficulty"`
	Nonce            int               `json:"nonce"`
	PoWValid         bool              `json:"powValid"`
	Timestamp        string            `json:"timestamp"`
	TransactionCount int               `json:"transactionCount"`
	Transactions     []TransactionView `json:"transactions"`
}

type TransactionView struct {
	ID      string       `json:"id"`
	Fee     int          `json:"fee"`
	Inputs  []InputView  `json:"inputs"`
	Outputs []OutputView `json:"outputs"`
}

type InputView struct {
	TxID   string `json:"txid"`
	Out    int    `json:"out"`
	Source string `json:"source"`
}

type OutputView struct {
	To    string `json:"to"`
	Value int    `json:"value"`
}

type CommandResult struct {
	Command  string `json:"command"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

type NodeStatus struct {
	Address      string   `json:"address"`
	MinerAddress string   `json:"minerAddress"`
	Peers        []string `json:"peers"`
	Height       int      `json:"height"`
	Running      bool     `json:"running"`
}
