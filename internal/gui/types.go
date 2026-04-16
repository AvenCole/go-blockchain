package gui

type DashboardData struct {
	Height         int              `json:"height"`
	LatestHash     string           `json:"latestHash"`
	MerkleRoot     string           `json:"merkleRoot"`
	Difficulty     int              `json:"difficulty"`
	Nonce          int              `json:"nonce"`
	PendingTxCount int              `json:"pendingTxCount"`
	WalletCount    int              `json:"walletCount"`
	DataDir        string           `json:"dataDir"`
	NetworkMode    string           `json:"networkMode"`
	LastReorg      *ReorgStatusView `json:"lastReorg,omitempty"`
	RecentEvents   []ChainEventView `json:"recentEvents"`
}

type ReorgStatusView struct {
	Timestamp             string `json:"timestamp"`
	OldHeight             int    `json:"oldHeight"`
	NewHeight             int    `json:"newHeight"`
	OldTip                string `json:"oldTip"`
	NewTip                string `json:"newTip"`
	RestoredTxCount       int    `json:"restoredTxCount"`
	DroppedConfirmedCount int    `json:"droppedConfirmedCount"`
}

type ChainEventView struct {
	Timestamp             string `json:"timestamp"`
	Kind                  string `json:"kind"`
	Summary               string `json:"summary"`
	OldHeight             int    `json:"oldHeight"`
	NewHeight             int    `json:"newHeight"`
	OldTip                string `json:"oldTip"`
	NewTip                string `json:"newTip"`
	RestoredTxCount       int    `json:"restoredTxCount"`
	DroppedConfirmedCount int    `json:"droppedConfirmedCount"`
}

type WalletView struct {
	Address       string `json:"address"`
	Balance       int    `json:"balance"`
	LockingScript string `json:"lockingScript"`
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
	ID           string       `json:"id"`
	Version      int          `json:"version"`
	Fee          int          `json:"fee"`
	UsesScriptVM bool         `json:"usesScriptVM"`
	Inputs       []InputView  `json:"inputs"`
	Outputs      []OutputView `json:"outputs"`
}

type InputView struct {
	TxID      string `json:"txid"`
	Out       int    `json:"out"`
	Source    string `json:"source"`
	ScriptSig string `json:"scriptSig"`
}

type OutputView struct {
	To           string `json:"to"`
	Value        int    `json:"value"`
	ScriptPubKey string `json:"scriptPubKey"`
}

type MultiSigOutputView struct {
	TxID         string   `json:"txid"`
	Out          int      `json:"out"`
	Value        int      `json:"value"`
	Required     int      `json:"required"`
	Participants []string `json:"participants"`
	ScriptPubKey string   `json:"scriptPubKey"`
}

type CommandResult struct {
	Command  string `json:"command"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

type NodeStatus struct {
	Address      string           `json:"address"`
	MinerAddress string           `json:"minerAddress"`
	Peers        []string         `json:"peers"`
	Initialized  bool             `json:"initialized"`
	Height       int              `json:"height"`
	TipHash      string           `json:"tipHash"`
	MempoolCount int              `json:"mempoolCount"`
	Running      bool             `json:"running"`
	OrphanCount  int              `json:"orphanCount"`
	LastReorg    *ReorgStatusView `json:"lastReorg,omitempty"`
	RecentEvents []NodeEventView  `json:"recentEvents"`
}

type NodeEventView struct {
	Timestamp string `json:"timestamp"`
	Kind      string `json:"kind"`
	Detail    string `json:"detail"`
}

type NetworkDemoResult struct {
	SourceNode      string `json:"sourceNode"`
	PeerNode        string `json:"peerNode"`
	MinerAddress    string `json:"minerAddress"`
	ReceiverAddress string `json:"receiverAddress"`
	TxID            string `json:"txid"`
	BlockHash       string `json:"blockHash"`
	PeerHeight      int    `json:"peerHeight"`
	TipAnnounced    bool   `json:"tipAnnounced"`
}

type NetworkReorgDemoResult struct {
	SourceNode          string `json:"sourceNode"`
	PeerNode            string `json:"peerNode"`
	MinerAddress        string `json:"minerAddress"`
	ReceiverAddress     string `json:"receiverAddress"`
	OriginalBlockHash   string `json:"originalBlockHash"`
	OriginalBlockHeight int    `json:"originalBlockHeight"`
	ReorgTxID           string `json:"reorgTxID"`
	Restored            bool   `json:"restored"`
	SourceOldHeight     int    `json:"sourceOldHeight"`
	SourceNewHeight     int    `json:"sourceNewHeight"`
	PeerHeight          int    `json:"peerHeight"`
	PeerReorged         bool   `json:"peerReorged"`
}

type NetworkPartitionDemoResult struct {
	SourceNode         string `json:"sourceNode"`
	PeerNode           string `json:"peerNode"`
	ForkNode           string `json:"forkNode"`
	MinerAddress       string `json:"minerAddress"`
	ReceiverAddress    string `json:"receiverAddress"`
	ConfirmedTxID      string `json:"confirmedTxID"`
	OldConfirmedHeight int    `json:"oldConfirmedHeight"`
	ForkHeight         int    `json:"forkHeight"`
	FinalTipHash       string `json:"finalTipHash"`
	Restored           bool   `json:"restored"`
	AllConverged       bool   `json:"allConverged"`
}

type NetworkOperationProgress struct {
	Operation   string `json:"operation"`
	Status      string `json:"status"`
	Phase       string `json:"phase"`
	Message     string `json:"message"`
	CurrentStep int    `json:"currentStep"`
	TotalSteps  int    `json:"totalSteps"`
	StartedAt   string `json:"startedAt"`
	FinishedAt  string `json:"finishedAt,omitempty"`
	ElapsedMS   int64  `json:"elapsedMs"`
	Error       string `json:"error,omitempty"`
	Summary     string `json:"summary,omitempty"`
}
