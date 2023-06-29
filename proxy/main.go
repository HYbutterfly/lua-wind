package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
)

const socketPath = "/tmp/windproxy.sock"

func main() {
	cleanup()
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Listening on Unix Domain Socket %s\n", socketPath)
	for {
		client, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleClient(client)
	}
}

func handleClient(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	for {
		pkgLenBuf := make([]byte, 2)
		if _, err := io.ReadFull(reader, pkgLenBuf); err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading package length from socket: %v\n", err)
			}
			return
		}
		pkgLen := binary.BigEndian.Uint16(pkgLenBuf)
		pkgBuf := make([]byte, pkgLen)
		if _, err := io.ReadFull(reader, pkgBuf); err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading package body from socket: %v\n", err)
			}
			return
		}
		fmt.Printf("Received package with length %v: %v\n", pkgLen, string(pkgBuf))
		go handleRequest(c, pkgBuf)
	}
}

func handleRequest(c net.Conn, message []byte) {
	var data []interface{}
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		fmt.Println(err)
	}

	session := data[0].(string)
	cmd := data[1].(string)
	params := data[2].(map[string]interface{})

	if cmd == "http_request" {
		body, header, err := do_http_request(params)
		if err != nil {
			response(c, HttpRequestResponse{
				Session: session,
				Error:   err.Error(),
			})
		} else {
			response(c, HttpRequestResponse{
				Session: session,
				Header:  header,
				Body:    body,
			})
		}
	}
}

func do_http_request(params map[string]interface{}) (string, http.Header, error) {
	method := params["method"].(string)
	url := params["url"].(string)
	body := params["body"].(string)
	header := params["header"].(map[string]interface{})

	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value.(string))
	}

	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	} else {
		return string(result), resp.Header, nil
	}
}

func response(client net.Conn, r interface{}) {
	response, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = client.Write(string_pack(response))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Sent message to client: %s\n", string(response))
}

func cleanup() {
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.RemoveAll(socketPath); err != nil {
			fmt.Println(err)
		}
	}
}

// 大端2字节
func string_pack(data []byte) []byte {
	var length int = len(data)
	if length > (int(math.Pow(2, 16)) - 1) {
		panic(fmt.Sprintf("Value %d is out of range for uint16", length))
	}
	headerBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(headerBytes, uint16(length))
	newData := append(headerBytes, data...)
	return newData
}

type HttpRequestResponse struct {
	Session string      `json:"session"`
	Error   string      `json:"error"`
	Header  http.Header `json:"header"`
	Body    string      `json:"body"`
}
