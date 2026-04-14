export namespace gui {
	
	export class OutputView {
	    to: string;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new OutputView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.to = source["to"];
	        this.value = source["value"];
	    }
	}
	export class InputView {
	    txid: string;
	    out: number;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new InputView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.txid = source["txid"];
	        this.out = source["out"];
	        this.source = source["source"];
	    }
	}
	export class TransactionView {
	    id: string;
	    fee: number;
	    inputs: InputView[];
	    outputs: OutputView[];
	
	    static createFrom(source: any = {}) {
	        return new TransactionView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fee = source["fee"];
	        this.inputs = this.convertValues(source["inputs"], InputView);
	        this.outputs = this.convertValues(source["outputs"], OutputView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BlockView {
	    height: number;
	    hash: string;
	    prevHash: string;
	    merkleRoot: string;
	    difficulty: number;
	    nonce: number;
	    powValid: boolean;
	    timestamp: string;
	    transactionCount: number;
	    transactions: TransactionView[];
	
	    static createFrom(source: any = {}) {
	        return new BlockView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.height = source["height"];
	        this.hash = source["hash"];
	        this.prevHash = source["prevHash"];
	        this.merkleRoot = source["merkleRoot"];
	        this.difficulty = source["difficulty"];
	        this.nonce = source["nonce"];
	        this.powValid = source["powValid"];
	        this.timestamp = source["timestamp"];
	        this.transactionCount = source["transactionCount"];
	        this.transactions = this.convertValues(source["transactions"], TransactionView);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CommandResult {
	    command: string;
	    stdout: string;
	    stderr: string;
	    exitCode: number;
	
	    static createFrom(source: any = {}) {
	        return new CommandResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.stdout = source["stdout"];
	        this.stderr = source["stderr"];
	        this.exitCode = source["exitCode"];
	    }
	}
	export class DashboardData {
	    height: number;
	    latestHash: string;
	    merkleRoot: string;
	    difficulty: number;
	    nonce: number;
	    pendingTxCount: number;
	    walletCount: number;
	    dataDir: string;
	    networkMode: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.height = source["height"];
	        this.latestHash = source["latestHash"];
	        this.merkleRoot = source["merkleRoot"];
	        this.difficulty = source["difficulty"];
	        this.nonce = source["nonce"];
	        this.pendingTxCount = source["pendingTxCount"];
	        this.walletCount = source["walletCount"];
	        this.dataDir = source["dataDir"];
	        this.networkMode = source["networkMode"];
	    }
	}
	
	export class NodeStatus {
	    address: string;
	    minerAddress: string;
	    peers: string[];
	    height: number;
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NodeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.minerAddress = source["minerAddress"];
	        this.peers = source["peers"];
	        this.height = source["height"];
	        this.running = source["running"];
	    }
	}
	
	
	export class WalletView {
	    address: string;
	    balance: number;
	
	    static createFrom(source: any = {}) {
	        return new WalletView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.balance = source["balance"];
	    }
	}

}

