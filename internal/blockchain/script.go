package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"go-blockchain/internal/wallet"
)

type Opcode byte

const (
	OpPushData      Opcode = 0x00
	OpDup           Opcode = 0x76
	OpHash160       Opcode = 0xa9
	OpEqual         Opcode = 0x87
	OpEqualVerify   Opcode = 0x88
	OpCheckSig      Opcode = 0xac
	OpCheckMultiSig Opcode = 0xae
)

type ScriptCommand struct {
	Opcode Opcode
	Data   []byte
}

type Script struct {
	Commands []ScriptCommand
}

func NewPushData(data []byte) ScriptCommand {
	return ScriptCommand{
		Opcode: OpPushData,
		Data:   append([]byte(nil), data...),
	}
}

func NewScript(commands ...ScriptCommand) Script {
	script := Script{
		Commands: make([]ScriptCommand, len(commands)),
	}
	for i, command := range commands {
		script.Commands[i] = command.Clone()
	}
	return script
}

func NewP2PKHLockingScript(pubKeyHash []byte) Script {
	return NewScript(
		ScriptCommand{Opcode: OpDup},
		ScriptCommand{Opcode: OpHash160},
		NewPushData(pubKeyHash),
		ScriptCommand{Opcode: OpEqualVerify},
		ScriptCommand{Opcode: OpCheckSig},
	)
}

func NewP2PKLockingScript(pubKey []byte) Script {
	return NewScript(
		NewPushData(pubKey),
		ScriptCommand{Opcode: OpCheckSig},
	)
}

func NewMultiSigLockingScript(required int, pubKeys ...[]byte) Script {
	commands := []ScriptCommand{NewPushData(encodeInt(required))}
	for _, pubKey := range pubKeys {
		commands = append(commands, NewPushData(pubKey))
	}
	commands = append(commands, NewPushData(encodeInt(len(pubKeys))))
	commands = append(commands, ScriptCommand{Opcode: OpCheckMultiSig})
	return NewScript(commands...)
}

func NewP2PKHUnlockingScript(signature []byte, pubKey []byte) Script {
	return NewScript(
		NewPushData(signature),
		NewPushData(pubKey),
	)
}

func NewP2PKUnlockingScript(signature []byte) Script {
	return NewScript(NewPushData(signature))
}

func NewMultiSigUnlockingScript(signatures ...[]byte) Script {
	commands := make([]ScriptCommand, 0, len(signatures))
	for _, signature := range signatures {
		commands = append(commands, NewPushData(signature))
	}
	return NewScript(commands...)
}

func NewCoinbaseScript(data string) Script {
	return NewScript(NewPushData([]byte(data)))
}

func (cmd ScriptCommand) Clone() ScriptCommand {
	return ScriptCommand{
		Opcode: cmd.Opcode,
		Data:   append([]byte(nil), cmd.Data...),
	}
}

func (s Script) Clone() Script {
	return NewScript(s.Commands...)
}

func (s Script) IsEmpty() bool {
	return len(s.Commands) == 0
}

func (s Script) String() string {
	if len(s.Commands) == 0 {
		return "(empty)"
	}

	parts := make([]string, 0, len(s.Commands))
	for _, command := range s.Commands {
		if command.Opcode == OpPushData {
			parts = append(parts, fmt.Sprintf("DATA[%s]", shortHex(command.Data)))
			continue
		}
		parts = append(parts, command.Opcode.String())
	}

	return strings.Join(parts, " ")
}

func (op Opcode) String() string {
	switch op {
	case OpPushData:
		return "OP_PUSH"
	case OpDup:
		return "OP_DUP"
	case OpHash160:
		return "OP_HASH160"
	case OpEqual:
		return "OP_EQUAL"
	case OpEqualVerify:
		return "OP_EQUALVERIFY"
	case OpCheckSig:
		return "OP_CHECKSIG"
	case OpCheckMultiSig:
		return "OP_CHECKMULTISIG"
	default:
		return fmt.Sprintf("OP_UNKNOWN_%#x", byte(op))
	}
}

func ExtractP2PKHPubKeyHash(script Script) ([]byte, bool) {
	if len(script.Commands) != 5 {
		return nil, false
	}
	if script.Commands[0].Opcode != OpDup ||
		script.Commands[1].Opcode != OpHash160 ||
		script.Commands[2].Opcode != OpPushData ||
		script.Commands[3].Opcode != OpEqualVerify ||
		script.Commands[4].Opcode != OpCheckSig {
		return nil, false
	}

	return append([]byte(nil), script.Commands[2].Data...), true
}

func ExtractP2PKPubKey(script Script) ([]byte, bool) {
	if len(script.Commands) != 2 {
		return nil, false
	}
	if script.Commands[0].Opcode != OpPushData || script.Commands[1].Opcode != OpCheckSig {
		return nil, false
	}
	return append([]byte(nil), script.Commands[0].Data...), true
}

func ExtractMultiSigPubKeys(script Script) (int, [][]byte, bool) {
	if len(script.Commands) < 4 {
		return 0, nil, false
	}
	if script.Commands[len(script.Commands)-1].Opcode != OpCheckMultiSig {
		return 0, nil, false
	}
	if script.Commands[0].Opcode != OpPushData || script.Commands[len(script.Commands)-2].Opcode != OpPushData {
		return 0, nil, false
	}

	required, ok := decodeInt(script.Commands[0].Data)
	if !ok {
		return 0, nil, false
	}
	total, ok := decodeInt(script.Commands[len(script.Commands)-2].Data)
	if !ok || total <= 0 || total != len(script.Commands)-3 {
		return 0, nil, false
	}

	pubKeys := make([][]byte, 0, total)
	for i := 1; i <= total; i++ {
		if script.Commands[i].Opcode != OpPushData {
			return 0, nil, false
		}
		pubKeys = append(pubKeys, append([]byte(nil), script.Commands[i].Data...))
	}
	if required <= 0 || required > len(pubKeys) {
		return 0, nil, false
	}
	return required, pubKeys, true
}

func ExtractP2PKHUnlockingData(script Script) ([]byte, []byte, bool) {
	if len(script.Commands) != 2 {
		return nil, nil, false
	}
	if script.Commands[0].Opcode != OpPushData || script.Commands[1].Opcode != OpPushData {
		return nil, nil, false
	}

	return append([]byte(nil), script.Commands[0].Data...),
		append([]byte(nil), script.Commands[1].Data...),
		true
}

func ExtractP2PKSignature(script Script) ([]byte, bool) {
	if len(script.Commands) != 1 {
		return nil, false
	}
	if script.Commands[0].Opcode != OpPushData {
		return nil, false
	}
	return append([]byte(nil), script.Commands[0].Data...), true
}

func ExtractMultiSigSignatures(script Script) ([][]byte, bool) {
	if len(script.Commands) == 0 {
		return nil, false
	}
	signatures := make([][]byte, 0, len(script.Commands))
	for _, command := range script.Commands {
		if command.Opcode != OpPushData {
			return nil, false
		}
		signatures = append(signatures, append([]byte(nil), command.Data...))
	}
	return signatures, true
}

func VerifyScripts(unlocking Script, locking Script, digest []byte) bool {
	engine := newScriptEngine(digest)
	if !engine.Execute(unlocking) {
		return false
	}
	if !engine.Execute(locking) {
		return false
	}
	return engine.Result()
}

type scriptEngine struct {
	stack  [][]byte
	digest []byte
}

func newScriptEngine(digest []byte) *scriptEngine {
	return &scriptEngine{
		stack:  make([][]byte, 0, 8),
		digest: append([]byte(nil), digest...),
	}
}

func (e *scriptEngine) Execute(script Script) bool {
	for _, command := range script.Commands {
		switch command.Opcode {
		case OpPushData:
			e.push(command.Data)
		case OpDup:
			if len(e.stack) < 1 {
				return false
			}
			e.push(e.stack[len(e.stack)-1])
		case OpHash160:
			top, ok := e.pop()
			if !ok {
				return false
			}
			e.push(wallet.HashPublicKey(top))
		case OpEqual:
			right, left, ok := e.pop2()
			if !ok {
				return false
			}
			e.push(encodeBool(bytes.Equal(left, right)))
		case OpEqualVerify:
			right, left, ok := e.pop2()
			if !ok {
				return false
			}
			if !bytes.Equal(left, right) {
				return false
			}
		case OpCheckSig:
			pubKey, signature, ok := e.popPubKeyAndSignature()
			if !ok {
				return false
			}
			e.push(encodeBool(verifyECDSASignature(pubKey, signature, e.digest)))
		case OpCheckMultiSig:
			if !e.checkMultiSig() {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func (e *scriptEngine) Result() bool {
	top, ok := e.pop()
	if !ok {
		return false
	}
	return decodeBool(top)
}

func (e *scriptEngine) popPubKeyAndSignature() ([]byte, []byte, bool) {
	pubKey, ok := e.pop()
	if !ok {
		return nil, nil, false
	}
	signature, ok := e.pop()
	if !ok {
		return nil, nil, false
	}
	return pubKey, signature, true
}

func (e *scriptEngine) checkMultiSig() bool {
	totalBytes, ok := e.pop()
	if !ok {
		return false
	}
	total, ok := decodeInt(totalBytes)
	if !ok || total <= 0 || len(e.stack) < total+1 {
		return false
	}

	pubKeys := make([][]byte, total)
	for i := total - 1; i >= 0; i-- {
		pubKey, ok := e.pop()
		if !ok {
			return false
		}
		pubKeys[i] = pubKey
	}

	requiredBytes, ok := e.pop()
	if !ok {
		return false
	}
	required, ok := decodeInt(requiredBytes)
	if !ok || required <= 0 || required > total || len(e.stack) < required {
		return false
	}

	signatures := make([][]byte, required)
	for i := required - 1; i >= 0; i-- {
		signature, ok := e.pop()
		if !ok {
			return false
		}
		signatures[i] = signature
	}

	for i := 0; i < required; i++ {
		if !verifyECDSASignature(pubKeys[i], signatures[i], e.digest) {
			e.push(encodeBool(false))
			return true
		}
	}

	e.push(encodeBool(true))
	return true
}

func (e *scriptEngine) push(value []byte) {
	e.stack = append(e.stack, append([]byte(nil), value...))
}

func (e *scriptEngine) pop() ([]byte, bool) {
	if len(e.stack) == 0 {
		return nil, false
	}
	last := len(e.stack) - 1
	value := append([]byte(nil), e.stack[last]...)
	e.stack = e.stack[:last]
	return value, true
}

func (e *scriptEngine) pop2() ([]byte, []byte, bool) {
	right, ok := e.pop()
	if !ok {
		return nil, nil, false
	}
	left, ok := e.pop()
	if !ok {
		return nil, nil, false
	}
	return right, left, true
}

func verifyECDSASignature(pubKeyBytes []byte, signature []byte, digest []byte) bool {
	if len(signature) == 0 || len(signature)%2 != 0 {
		return false
	}

	x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
	if x == nil || y == nil {
		return false
	}

	r := big.Int{}
	s := big.Int{}
	r.SetBytes(signature[:len(signature)/2])
	s.SetBytes(signature[len(signature)/2:])

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return ecdsa.Verify(&publicKey, digest, &r, &s)
}

func encodeBool(value bool) []byte {
	if value {
		return []byte{1}
	}
	return []byte{}
}

func decodeBool(value []byte) bool {
	for _, b := range value {
		if b != 0 {
			return true
		}
	}
	return false
}

func shortHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	text := hex.EncodeToString(data)
	if len(text) <= 24 {
		return text
	}
	return text[:12] + "..." + text[len(text)-12:]
}

func encodeInt(value int) []byte {
	if value < 0 || value > 255 {
		return []byte{}
	}
	return []byte{byte(value)}
}

func decodeInt(data []byte) (int, bool) {
	if len(data) != 1 {
		return 0, false
	}
	return int(data[0]), true
}
