package rpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	ErrParse    = -32700
	ErrInvalid  = -32600
	ErrMethod   = -32601
	ErrInternal = -32603
	// Business errors
	ErrSkillNotFound   = -32001
	ErrPluginNotFound  = -32002
	ErrPathInvalid     = -32003
	ErrTranslationFail = -32004
)

type MethodFn func(params json.RawMessage) (any, error)

type Handler struct {
	mu       sync.RWMutex
	methods  map[string]MethodFn
	notifyCh chan map[string]any
}

func NewHandler(notifyCh chan map[string]any) *Handler {
	return &Handler{
		methods:  make(map[string]MethodFn),
		notifyCh: notifyCh,
	}
}

func (h *Handler) Register(method string, fn MethodFn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.methods[method] = fn
}

func (h *Handler) Notify(method string, params map[string]any) {
	if h.notifyCh != nil {
		h.notifyCh <- map[string]any{
			"method": method,
			"params": params,
		}
	}
}

func (h *Handler) handleRaw(line []byte) []byte {
	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		return errorResponse(nil, ErrParse, "invalid JSON")
	}
	if req.JSONRPC != "2.0" {
		return errorResponse(req.ID, ErrInvalid, "jsonrpc must be 2.0")
	}
	// Notification (no id) — ignore from client
	if req.ID == nil {
		return nil
	}
	h.mu.RLock()
	fn, ok := h.methods[req.Method]
	h.mu.RUnlock()
	if !ok {
		return errorResponse(req.ID, ErrMethod, fmt.Sprintf("unknown method: %s", req.Method))
	}
	result, err := fn(req.Params)
	if err != nil {
		return errorResponse(req.ID, ErrInternal, err.Error())
	}
	return successResponse(req.ID, result)
}

func errorResponse(id any, code int, msg string) []byte {
	r := Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
	b, _ := json.Marshal(r)
	return b
}

func successResponse(id any, result any) []byte {
	r := Response{JSONRPC: "2.0", ID: id, Result: result}
	b, _ := json.Marshal(r)
	return b
}

// RunLoop reads JSON-RPC lines from stdin, writes responses to stdout.
func (h *Handler) RunLoop() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	// Send ready notification
	fmt.Println(`{"jsonrpc":"2.0","method":"ready","params":{}}`)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		resp := h.handleRaw(line)
		if resp != nil {
			os.Stdout.Write(resp)
			os.Stdout.Write([]byte{'\n'})
		}
		// Drain pending notifications
		for {
			select {
			case n := <-h.notifyCh:
				b, _ := json.Marshal(map[string]any{
					"jsonrpc": "2.0",
					"method":  n["method"],
					"params":  n["params"],
				})
				os.Stdout.Write(b)
				os.Stdout.Write([]byte{'\n'})
			default:
				goto doneNotify
			}
		}
	doneNotify:
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "rpc: scan error: %v\n", err)
	}
}
