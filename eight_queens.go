package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/net/websocket"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.Handle("/compute", websocket.Handler(compute))

	log.Print("Server runing on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func compute(ws *websocket.Conn) {
	defer ws.Close()

	var recMsg string
	err := websocket.Message.Receive(ws, &recMsg)
	if err != nil {
		log.Print("Receiving error")
		return
	}

	size, err := strconv.Atoi(recMsg)
	if err != nil {
		log.Printf("%s is not a int", recMsg)
	}
	if size < 4 || size > 15 {
		log.Print("Size out of range [4; 15]")
		return
	}

	board := newBoard(size)
	var wg sync.WaitGroup

	setQueen(board, 0, &wg, newConnection(ws))
	wg.Wait()
}

type chessBoard struct {
	size int
	data []int
}

func newBoard(size int) *chessBoard {
	return &chessBoard{size, make([]int, size)}
}

func (board *chessBoard) Copy() *chessBoard {
	boardCopy := make([]int, board.size)
	copy(boardCopy, board.data)
	return &chessBoard{board.size, boardCopy}
}

type connection struct {
	m  *sync.Mutex
	ws *websocket.Conn
}

func newConnection(ws *websocket.Conn) *connection {
	return &connection{new(sync.Mutex), ws}
}

func (conn *connection) Send(board []int) {
	conn.m.Lock()
	websocket.JSON.Send(conn.ws, board)
	conn.m.Unlock()
}

func setQueen(board *chessBoard, depth int, wg *sync.WaitGroup, conn *connection) {
	switch depth {
	case 1:
		defer wg.Done()
	case board.size:
		conn.Send(board.data)
		return
	}

	for i := 0; i < board.size; i++ {
		if check(board, depth, i) {
			board.data[depth] = i

			if depth == 0 {
				wg.Add(1)
				go setQueen(board.Copy(), depth+1, wg, conn)
			} else {
				setQueen(board, depth+1, wg, conn)
			}
		}
	}
}

func check(board *chessBoard, depth, val int) bool {
	for i := 0; i < depth; i++ {
		cur := board.data[i]
		if cur == val || cur == val-depth+i || cur == val+depth-i {
			return false
		}
	}
	return true
}
