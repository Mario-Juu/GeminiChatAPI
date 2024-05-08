package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"nhooyr.io/websocket"
)

type Message struct{
	User bool `json:"user"`
	Content string `json:"content"`
	SentAt string `json:"sentAt"`
}

func iaResp(msg string) (genai.Content, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey("AIzaSyAsYA1lW1zUN5mCSyYZGJS95iGGReGzpR8"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-pro")
	resp, err := model.GenerateContent(context.Background(), genai.Text(msg))
	if err != nil {
		log.Fatal(err)
	}
	return *resp.Candidates[0].Content, nil

}


func wsHandler(w http.ResponseWriter, r *http.Request){
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil{
		log.Println(err)
		return
	}

	for{
		_, data, err := conn.Read(r.Context())
		if err != nil{
			log.Print("Encerrando a conexão.")
			break
		}

		var msgRec Message
		json.Unmarshal(data, &msgRec)
		MessageUser := Message{User: true, Content: msgRec.Content, SentAt: time.Now().Format("2006-01-02 15:04:05")}
		broadcast(MessageUser, conn)

		resp, err := iaResp(msgRec.Content)
		if err != nil{
			log.Print(err)
		}

		
		log.Println(msgRec.Content)

		content, _ := json.Marshal(resp.Parts[0])
		stringContent := string(content)
		
		
		log.Println(stringContent)

		stringContent = stringContent[1 : len(stringContent)-1]
		msg := Message{User: false, Content: stringContent, SentAt: time.Now().Format("2006-01-02 15:04:05")}
		broadcast(msg, conn)
	}
	
}

func broadcast(msg Message, conn *websocket.Conn){
	data, _ := json.Marshal(msg)
	conn.Write(context.Background(), websocket.MessageText, data)
}

func main(){
	
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/", http.FileServer(http.Dir("./public")))


	log.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}